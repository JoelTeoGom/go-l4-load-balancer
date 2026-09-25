package apiserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/etcd"
	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/registry"
)

type APIServer struct {
	Registry   *registry.Registry
	EventQueue chan Event
	Db         etcd.Etcd
}

func NewAPIServer(registry *registry.Registry, storage etcd.Etcd) *APIServer {
	return &APIServer{
		Db:         storage,
		Registry:   registry,
		EventQueue: make(chan Event, 100),
	}
}

func (cp *APIServer) StartAPIServer(ctx context.Context, address string) error {
	http.HandleFunc("/register-node", cp.registerNodeHandler)
	http.HandleFunc("/unregister-node", cp.unregisterNodeHandler)
	http.HandleFunc("/list-nodes", cp.listNodesHandler)
	http.HandleFunc("/health", cp.healthHandler)

	http.HandleFunc("GET /list/{id}", cp.ListNodeDataHandler)
	http.HandleFunc("/create-service", cp.CreateServiceHandler)
	http.HandleFunc("/create-pod", cp.CreatePodHandler)
	http.HandleFunc("/watch-node", cp.WatchNodeHandler)

	fmt.Println("Listening Control Plane: ", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
