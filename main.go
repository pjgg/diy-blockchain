package main

import (
	"context"

	"diy.blockchain.org/m/api"
	"diy.blockchain.org/m/configuration"
)

func main() {
	ctx := context.Background()
	configuration.LoadConfig(ctx, "config.yaml")
	api.Start(ctx, &configuration.InstanceConfig)
}
