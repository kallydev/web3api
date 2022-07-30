package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eko/gocache/v3/store"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kallydev/web3api/common/contract/erc20"
	"github.com/kallydev/web3api/common/contract/uniswap"
	"github.com/kallydev/web3api/common/ethereum"
	"github.com/kallydev/web3api/common/thegraph"
	"github.com/kallydev/web3api/model"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

var ErrorUnsupportedPlatform = errors.New("unsupported platform")

type GetPoolRequest struct {
	Network         string `param:"network" validate:"network"`
	Platform        string `param:"platform"`
	ContractAddress string `param:"contract_address" validate:"ethereum_address"`
}

type GetPoolResponse struct {
	ContractAddress common.MixedcaseAddress `json:"contract_address"`
	TokenLeft       *model.Token            `json:"token_left"`
	TokenRight      *model.Token            `json:"token_right"`
	Fee             string                  `json:"fee"`
	Metrics         []model.PoolMetric      `json:"metrics"`
}

func (i *internal) GetPool(c echo.Context) error {
	var request GetPoolRequest

	if err := c.Bind(&request); err != nil {
		return err
	}

	if err := c.Validate(&request); err != nil {
		return err
	}

	if request.Platform != "uniswap" {
		return ErrorUnsupportedPlatform
	}

	value, err := i.cache.Get(c.Request().Context(), c.Request().URL.Path)
	if err != nil && !errors.Is(err, store.NotFound{}) {
		return err
	}

	if errors.Is(err, store.NotFound{}) {
		thegraphClient := thegraph.New()

		zap.L().Info(request.ContractAddress)

		uniswapPoolDayDataList, err := thegraphClient.GetUniswapPoolDayDataList(c.Request().Context(), request.Network, common.HexToAddress(request.ContractAddress))
		if err != nil {
			return err
		}

		if value, err = json.Marshal(uniswapPoolDayDataList); err != nil {
			return err
		}

		if err := i.cache.Set(c.Request().Context(), c.Request().URL.Path, value, store.WithExpiration(time.Minute)); err != nil {
			return err
		}
	}

	var uniswapPoolDayDataList []thegraph.UniswapPoolDayData
	if err := json.NewDecoder(bytes.NewReader(value)).Decode(&uniswapPoolDayDataList); err != nil {
		return err
	}

	var poolMetrics []model.PoolMetric

	for _, data := range uniswapPoolDayDataList {
		poolMetric := model.PoolMetric{
			Timestamp:       time.Unix(int64(data.Date), 0),
			ContractAddress: common.NewMixedcaseAddress(common.HexToAddress(request.ContractAddress)),
		}

		if poolMetric.TotalValueLocked, err = decimal.NewFromString(string(data.TotalValueLocked)); err != nil {
			return err
		}

		if poolMetric.Volume, err = decimal.NewFromString(string(data.Volume)); err != nil {
			return err
		}

		if poolMetric.Fee, err = decimal.NewFromString(string(data.Fees)); err != nil {
			return err
		}

		if poolMetric.TotalValueLocked.Cmp(decimal.Zero) == 0 || poolMetric.Fee.Cmp(decimal.Zero) == 0 {
			poolMetric.AnnualPercentageRate = "0%"
		} else {
			// 1 / TVL * Fee * 365 / 1
			poolMetric.AnnualPercentageRate = fmt.Sprintf("%s%%", decimal.NewFromInt(1).Div(poolMetric.TotalValueLocked).Mul(poolMetric.Fee).Mul(decimal.NewFromInt(365).Div(decimal.NewFromInt(1))).Shift(2).StringFixedBank(2))
		}

		poolMetrics = append(poolMetrics, poolMetric)
	}

	ethereumClient, exists := i.ethereumClientMap[request.Network]
	if !exists {
		return ethereum.ErrorUnsupportedNetwork
	}

	poolAddress := common.HexToAddress(request.ContractAddress)

	response := GetPoolResponse{
		ContractAddress: common.NewMixedcaseAddress(poolAddress),
		Metrics:         poolMetrics,
	}

	pool, err := i.buildPool(c.Request().Context(), poolAddress, ethereumClient)
	if err != nil {
		return err
	}

	response.Fee = fmt.Sprintf("%s%%", decimal.NewFromBigInt(pool.Fee, 0).Shift(-4).String())

	if response.TokenLeft, err = i.buildToken(c.Request().Context(), pool.TokenLeft, ethereumClient); err != nil {
		return err
	}

	if response.TokenRight, err = i.buildToken(c.Request().Context(), pool.TokenRight, ethereumClient); err != nil {
		return err
	}

	return c.JSONPretty(http.StatusOK, &response, "\x20\x20")
}

func (i *internal) buildPool(ctx context.Context, contractAddress common.Address, ethereumClient *ethclient.Client) (*model.Pool, error) {
	value, err := i.cache.Get(ctx, strings.ToLower(contractAddress.String()))
	if err != nil && !errors.Is(err, store.NotFound{}) {
		return nil, err
	}

	if err == nil {
		var pool model.Pool

		if err := json.Unmarshal(value, &pool); err != nil {
			return nil, err
		}

		return &pool, nil
	}

	poolV3Contract, err := uniswap.NewPoolV3(contractAddress, ethereumClient)
	if err != nil {
		return nil, err
	}

	address := common.NewMixedcaseAddress(contractAddress)

	pool := model.Pool{
		ContractAddress: common.HexToAddress(address.String()),
	}

	if pool.TokenLeft, err = poolV3Contract.Token0(&bind.CallOpts{}); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	if pool.TokenRight, err = poolV3Contract.Token1(&bind.CallOpts{}); err != nil {
		return nil, err
	}

	if pool.Fee, err = poolV3Contract.Fee(&bind.CallOpts{}); err != nil {
		return nil, err
	}

	if value, err = json.Marshal(pool); err != nil {
		return nil, err
	}

	if err := i.cache.Set(context.Background(), strings.ToLower(contractAddress.String()), value); err != nil {
		return nil, err
	}

	return &pool, nil
}

func (i *internal) buildToken(ctx context.Context, contractAddress common.Address, ethereumClient *ethclient.Client) (*model.Token, error) {
	value, err := i.cache.Get(ctx, strings.ToLower(contractAddress.String()))
	if err != nil && !errors.Is(err, store.NotFound{}) {
		return nil, err
	}

	if err == nil {
		var token model.Token

		if err := json.Unmarshal(value, &token); err != nil {
			return nil, err
		}

		return &token, nil
	}

	tokenLeftContract, err := erc20.NewERC20(contractAddress, ethereumClient)
	if err != nil {
		return nil, err
	}

	address := common.NewMixedcaseAddress(contractAddress)

	token := model.Token{
		ContractAddress: common.HexToAddress(address.String()),
	}

	if token.Name, err = tokenLeftContract.Name(&bind.CallOpts{}); err != nil {
		return nil, err
	}

	if token.Symbol, err = tokenLeftContract.Symbol(&bind.CallOpts{}); err != nil {
		return nil, err
	}

	if token.Decimals, err = tokenLeftContract.Decimals(&bind.CallOpts{}); err != nil {
		return nil, err
	}

	if value, err = json.Marshal(token); err != nil {
		return nil, err
	}

	if err := i.cache.Set(context.Background(), strings.ToLower(contractAddress.String()), value); err != nil {
		return nil, err
	}

	return &token, nil
}
