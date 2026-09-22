package main

import (
	"context"

	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/apiserver"
	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/config"
	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/etcd"
	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/loadbalancer"
	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/metrics"
	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/registry"
)

func main() {
	ctx := context.Background()
	cfg := config.NewConfig()

	etcd := etcd.NewEtcd()
	registry := registry.NewRegistry()
	metrics := metrics.NewMetrics(registry)
	controlPlane := apiserver.NewAPIServer(registry)
	dataPlane := loadbalancer.NewLoadBalancer(registry)

	go func() {
		err := controlPlane.StartAPIServer(ctx, cfg.CtrlPlaneAddress())
		if err != nil {
			panic(err)
		}
	}()

	go metrics.PrintMetrics(ctx)

	dataPlane.StartLoadBalancer(ctx, cfg.DataPlaneAddress())
}
