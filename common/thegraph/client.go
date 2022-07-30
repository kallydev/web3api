package thegraph

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hasura/go-graphql-client"
	"github.com/kallydev/web3api/common/ethereum"
)

const endpoint = "https://api.thegraph.com/subgraphs/name/"

type client struct {
	httpClient *http.Client
}

func (c *client) GetUniswapPoolDayDataList(ctx context.Context, network string, address common.Address) ([]UniswapPoolDayData, error) {
	var name string

	switch network {
	case ethereum.NetworkEthereum:
		name = "ianlapham/uniswap-v3-subgraph"
	case ethereum.NetworkPolygon:
		name = "ianlapham/uniswap-v3-polygon"
	default:
		return nil, ethereum.ErrorUnsupportedNetwork
	}

	graphqlClient := graphql.NewClient(fmt.Sprintf("%s%s", endpoint, name), c.httpClient)

	var query struct {
		PoolDayDataList []UniswapPoolDayData `graphql:"poolDayDatas(first: 1000, skip: $skip, where: {pool: $address, date_gt: $startTime}, orderBy: date, orderDirection: desc, subgraphError: allow)"`
	}

	if err := graphqlClient.Query(ctx, &query, map[string]any{
		"address":   graphql.String(strings.ToLower(address.String())),
		"startTime": graphql.Int(0),
		"skip":      graphql.Int(0),
	}, graphql.OperationName("poolDayDatas")); err != nil {
		return nil, err
	}

	return query.PoolDayDataList, nil
}

func New() *client {
	return &client{
		httpClient: http.DefaultClient,
	}
}
