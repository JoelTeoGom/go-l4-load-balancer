package main

import (
	"context"

	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/config"
	controlplane "github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/controlPlane"
	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/dataplane"
	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/registry"
)

func main() {
	ctx := context.Background()
	cfg := config.NewConfig()
	registry := registry.NewRegistry()
	controlPlane := controlplane.NewControlPlane(registry)
	dataPlane := dataplane.NewDataPlane(registry)

	go func() {
		err := controlPlane.StartControlPlane(ctx, cfg.Address())
		if err != nil {
			panic(err)
		}
	}()

	dataPlane.StartDataplane(ctx, cfg.Address())
}
