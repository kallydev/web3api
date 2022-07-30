package model

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type Pool struct {
	ContractAddress common.Address `json:"contract_address"`
	TokenLeft       common.Address `json:"token_left"`
	TokenRight      common.Address `json:"token_right"`
	Fee             *big.Int       `json:"fee"`
}

type Token struct {
	Name            string         `json:"name"`
	Symbol          string         `json:"symbol"`
	Decimals        uint8          `json:"decimals"`
	ContractAddress common.Address `json:"contract_address"`
}

type PoolMetric struct {
	Timestamp            time.Time               `json:"timestamp"`
	ContractAddress      common.MixedcaseAddress `json:"contract_address"`
	TotalValueLocked     decimal.Decimal         `json:"total_value_locked"`
	Volume               decimal.Decimal         `json:"volume"`
	Fee                  decimal.Decimal         `json:"fee"`
	AnnualPercentageRate string                  `json:"annual_percentage_rate"`
}
