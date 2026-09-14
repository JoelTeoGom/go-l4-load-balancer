package main

import (
	"context"

	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/config"
	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/controlplane"
	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/dataplane"
	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/metrics"
	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/registry"
)

func main() {
	ctx := context.Background()
	cfg := config.NewConfig()
	registry := registry.NewRegistry()
	metrics := metrics.NewMetrics(registry)
	controlPlane := controlplane.NewControlPlane(registry)
	dataPlane := dataplane.NewDataPlane(registry)

	go func() {
		err := controlPlane.StartControlPlane(ctx, cfg.Address())
		if err != nil {
			panic(err)
		}
	}()

	go metrics.PrintMetrics(ctx)

	dataPlane.StartDataplane(ctx, cfg.Address())
}
