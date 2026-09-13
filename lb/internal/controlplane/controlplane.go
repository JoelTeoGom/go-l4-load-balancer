package controlplane

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/registry"
)

type ControlPlane struct {
	registry *registry.Registry
}

func NewControlPlane(registry *registry.Registry) *ControlPlane {
	return &ControlPlane{
		registry: registry,
	}
}

func (cp *ControlPlane) StartControlPlane(ctx context.Context, address string) error {
	http.HandleFunc("/register-node", cp.registerNodeHandler)
	http.HandleFunc("/unregister-node", cp.unregisterNodeHandler)
	http.HandleFunc("/list-nodes", cp.listNodesHandler)
	http.HandleFunc("/health", cp.healthHandler)

	address = fmt.Sprintf("%s:9000", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
}
