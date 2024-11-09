package api

import (
	"context"
	"net/http"

	"diy.blockchain.org/m/configuration"
	"diy.blockchain.org/m/logger"
	"go.uber.org/zap"
)

func Start(ctx context.Context, configuration *configuration.Config) {
	http.HandleFunc("/health", HealthHandlerInstance().Health())
	http.HandleFunc("/transactions/new", BlockAndChainHandlerInstance().NewTransaction())
	http.HandleFunc("/mine", BlockAndChainHandlerInstance().MineBlock())
	http.HandleFunc("/chain", BlockAndChainHandlerInstance().GetChain())

	logger.Infof("Server started on port %s", configuration.HttpPort)
	logger.Fatal("Server didn't start.", zap.Error(http.ListenAndServe(":"+configuration.HttpPort, nil)))
}
