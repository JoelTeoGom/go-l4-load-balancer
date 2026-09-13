package controlplane

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/registry"
)

func (cp *ControlPlane) registerNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close() // Close the request body to avoid resource leaks

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Invalid remote address", http.StatusBadRequest)
		return
	}
	nodeID := fmt.Sprintf("Node-A")
	node := registry.NewNode(nodeID, host, registry.StatusActive, 0)
	if cp.registry.AddNode(node) == nil {
		http.Error(w, "Failed to register node", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Node registered successfully"))
}

func (cp *ControlPlane) unregisterNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// if !cp.registry.RemoveNode(r.Body) {
	// 	http.Error(w, "Failed to unregister node", http.StatusInternalServerError)
	// 	return
	// }

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Node unregistered successfully"))
}

func (cp *ControlPlane) listNodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodes := cp.registry.ListNodes()
	fmt.Println("Log List: ", nodes)
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte(fmt.Sprintf("List of nodes: %v", nodes)))
}

func (cp *ControlPlane) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Invalid remote address", http.StatusBadRequest)
		return
	}
	node := cp.registry.GetNoteByAddress(host)
	if node == nil {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}
	ctx := context.Background()
	cp.registry.EmitEvent(ctx, node, "CREATED")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Control plane is healthy"))
}
