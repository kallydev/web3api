package main

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/kallydev/web3api/common/ethereum"
	"github.com/kallydev/web3api/handler"
	"github.com/kallydev/web3api/middleware"
	"github.com/kallydev/web3api/validator"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	viper.AutomaticEnv()
}

func main() {
	logger, _ := zap.NewProduction()

	zap.ReplaceGlobals(logger)

	database, err := pgx.Connect(context.Background(), viper.GetString("POSTGRES_DSN"))
	if err != nil {
		zap.L().Fatal("failed to connect to the database", zap.Error(err))
	}

	defer func() {
		_ = database.Close(context.Background())
	}()

	httpServer := echo.New()

	httpServer.Use(middleware.ZapLogger(zap.L()))

	httpHandler, err := handler.New(database, &ethereum.Config{
		Network: &ethereum.Network{
			Ethereum: &ethereum.Endpoint{
				HTTP: viper.GetString("RPC_ETHEREUM_HTTP"),
			},
			Polygon: &ethereum.Endpoint{
				HTTP: viper.GetString("RPC_POLYGON_HTTP"),
			},
			Optimism: &ethereum.Endpoint{
				HTTP: viper.GetString("RPC_OPTIMISM_HTTP"),
			},
			Arbitrum: &ethereum.Endpoint{
				HTTP: viper.GetString("RPC_ARBITRUM_HTTP"),
			},
		},
	})
	if err != nil {
		zap.L().Fatal("failed to create the handler", zap.Error(err))
	}

	if httpServer.Validator, err = validator.New(); err != nil {
		zap.L().Fatal("failed to create the validator", zap.Error(err))
	}

	httpServer.HTTPErrorHandler = httpHandler.Error

	httpServer.GET("/pools/:platform/:network/:contract_address", httpHandler.GetPool)

	if err := httpServer.Start(":80"); err != nil {
		zap.L().Fatal("failed to start the http server", zap.Error(err))
	}
}
