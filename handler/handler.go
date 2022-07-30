package handler

import (
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/kallydev/web3api/common/ethereum"
)

type internal struct {
	cache             *cache.Cache[[]byte]
	database          *pgx.Conn
	ethereumClientMap map[string]*ethclient.Client
}

func New(database *pgx.Conn, ethereumConfig *ethereum.Config) (*internal, error) {
	handler := internal{
		database:          database,
		ethereumClientMap: make(map[string]*ethclient.Client),
	}

	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	handler.cache = cache.New[[]byte](store.NewRistretto(ristrettoCache))

	endpoints := map[string]string{
		ethereum.NetworkEthereum: ethereumConfig.Network.Ethereum.HTTP,
		ethereum.NetworkPolygon:  ethereumConfig.Network.Polygon.HTTP,
		ethereum.NetworkOptimism: ethereumConfig.Network.Optimism.HTTP,
		ethereum.NetworkArbitrum: ethereumConfig.Network.Arbitrum.HTTP,
	}

	for network, endpoint := range endpoints {
		if handler.ethereumClientMap[network], err = ethclient.Dial(endpoint); err != nil {
			return nil, err
		}
	}

	return &handler, nil
}
