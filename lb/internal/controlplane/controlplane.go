package controlplane

import (
	"context"
	"fmt"
	"net/http"
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
	//	http.HandleFunc("/health", healthHandler)
	//	http.HandleFunc("/watch-agent", watchAgentHandler)

	address := fmt.Sprintf("%s:9000", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
}

func (cp *ControlPlane) registerNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var node *registry.Node
	if cp.registry.AddNode(r.Body) == nil {
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

	if !cp.registry.RemoveNode(r.Body) {
		http.Error(w, "Failed to unregister node", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Node unregistered successfully"))
}

func (cp *ControlPlane) listNodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodes := cp.registry.ListNodes()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("List of nodes: %v", nodes)))
}

// address := fmt.Sprintf("%s:9000", address)
// ln, err := net.Listen("tcp", address)
// if err != nil {
// 	fmt.Println(err)
// 	return err
// }

// for {
// 	conn, err := ln.Accept()
// 	if err != nil {
// 		fmt.Println(err)
// 		continue
// 	}

// 	lb.ctrl.mu.Lock()
// 	lb.ctrl.nextNode++
// 	id := lb.ctrl.nextNode
// 	node := NewNode(id, conn)
// 	lb.ctrl.nodes = append(lb.ctrl.nodes, node)
// 	lb.ctrl.mu.Unlock()

// 	go lb.watchAgent(ctx, conn)
// }
// func (lb *ControlPlane) watchAgentHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Cache-Control", "no-cache")

// 	rc := http.NewResponseController(w)
// 	ticker := time.NewTicker(15 * time.Second)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case <-r.Context().Done():
// 			return
// 		case ev := <-lb.events:
// 			fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", ev.ID, ev.Name, ev.JSON)
// 			if err := rc.Flush(); err != nil {
// 				return
// 			}
// 		case <-ticker.C:
// 			fmt.Fprint(w, ": ping\n\n")
// 			if err := rc.Flush(); err != nil {
// 				return
// 			}
// 		}
// 	}
// }
