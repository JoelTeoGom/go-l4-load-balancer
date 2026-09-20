package apiserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JoelTeoGom/go-l4-load-balancer/orchestrator/internal/registry"
)

type APIServer struct {
	Registry   *registry.Registry
	EventQueue chan Event
}

func NewAPIServer(registry *registry.Registry) *APIServer {
	return &APIServer{
		Registry:   registry,
		EventQueue: make(chan Event, 100),
	}
}

func (cp *APIServer) StartAPIServer(ctx context.Context, address string) error {
	http.HandleFunc("/register-node", cp.registerNodeHandler)
	http.HandleFunc("/unregister-node", cp.unregisterNodeHandler)
	http.HandleFunc("/list-nodes", cp.listNodesHandler)
	http.HandleFunc("/health", cp.healthHandler)

	http.HandleFunc("/create-service", cp.CreateServiceHandler)
	http.HandleFunc("/create-pod", cp.CreatePodHandler)
	http.HandleFunc("/watch-node", cp.WatchNodeHandler)

	address = fmt.Sprintf("%s:9000", address)
	fmt.Println("Listening Control Plane: ", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
