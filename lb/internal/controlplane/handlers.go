package controlplane

import (
	"net"
	"net/http"
)

func (cp *ControlPlane) registerNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close() // Close the request body to avoid resource leaks

	host, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Invalid remote address", http.StatusBadRequest)
		return
	}
	// var node *registry.Node
	// node = registry.AddNode() // Create a new node with a unique ID and nil connection
	// if cp.registry.AddNode(node) == nil {
	// 	http.Error(w, "Failed to register node", http.StatusInternalServerError)
	// 	return
	// }

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

	//nodes := cp.registry.ListNodes()

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

	cp.registry.EmitEvent(node)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Control plane is healthy"))
}
