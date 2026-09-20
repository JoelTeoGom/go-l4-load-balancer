package main

import (
	"context"

	"github.com/JoelTeoGom/go-l4-load-balancer/orchestrator/internal/apiserver"
	"github.com/JoelTeoGom/go-l4-load-balancer/orchestrator/internal/config"
	"github.com/JoelTeoGom/go-l4-load-balancer/orchestrator/internal/loadbalancer"
	"github.com/JoelTeoGom/go-l4-load-balancer/orchestrator/internal/metrics"
	"github.com/JoelTeoGom/go-l4-load-balancer/orchestrator/internal/registry"
)

func main() {
	ctx := context.Background()
	cfg := config.NewConfig()
	registry := registry.NewRegistry()
	metrics := metrics.NewMetrics(registry)
	controlPlane := apiserver.NewAPIServer(registry)
	dataPlane := loadbalancer.NewLoadBalancer(registry)

	go func() {
		err := controlPlane.StartAPIServer(ctx, cfg.Address())
		if err != nil {
			panic(err)
		}
	}()

	go metrics.PrintMetrics(ctx)

	dataPlane.StartLoadBalancer(ctx, cfg.Address())
}
