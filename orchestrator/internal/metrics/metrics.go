package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/registry"
)

type Metrics struct {
	registry *registry.Registry
}

func NewMetrics(registry *registry.Registry) *Metrics {
	return &Metrics{
		registry: registry,
	}
}

func (m *Metrics) PrintMetrics(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			fmt.Println(m.registry.ListNodes())
		case <-ctx.Done():
			return
		}

	}
}
