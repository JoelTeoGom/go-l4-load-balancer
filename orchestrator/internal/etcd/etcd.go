package etcd

import (
	"context"

	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/data"
)

type Etcd struct {
	path string
}

func NewEtcd() *Etcd {
	return &Etcd{}
}

// We use this to reload node settings
func (r *Etcd) LoadNodeSettings(nodeID string) {

	return
}

// We Use this To store Node Settings
func (r *Etcd) CreateNode(ctx context.Context, nodeID string, serviceNames []string) error {
	return nil
}

func (r *Etcd) CreateService(ctx context.Context, serviceData data.Service) error {
	return nil
}
