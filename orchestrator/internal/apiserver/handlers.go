package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/registry"
)

func (cp *APIServer) registerNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close() // Close the request body to avoid resource leaks

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req RegisterNodeRequest
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusInternalServerError)
		return
	}

	node := registry.NewNode(req.NodeID, req.NodeIP, registry.StatusActive, 0, nil)
	if cp.Registry.AddNode(node) == nil {
		http.Error(w, "Failed to register node", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Node registered successfully"))
}

func (cp *APIServer) unregisterNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// if !cp.Registry.RemoveNode(r.Body) {
	// 	http.Error(w, "Failed to unregister node", http.StatusInternalServerError)
	// 	return
	// }

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Node unregistered successfully"))
}

func (cp *APIServer) listNodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodes := cp.Registry.ListNodes()
	fmt.Println("Log List: ", nodes)
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte(fmt.Sprintf("List of nodes: %v", nodes)))
}

func (cp *APIServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Invalid remote address", http.StatusBadRequest)
		return
	}
	node := cp.Registry.GetNodeByAddress(host)
	if node == nil {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}
	ctx := context.Background()
	cp.Registry.EmitEvent(ctx, node, "CREATED")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Control plane is healthy"))
}

func (cp *APIServer) CreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close() // Close the request body to avoid resource leaks

	event := Event{
		ID:      time.Now().String(),
		action:  ActionCreateService,
		Payload: "ServiceName",
	}

	//try to push if we dont have space we discard until next (we also use queue to rate limit)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		http.Error(w, "Too many request", http.StatusTooManyRequests)
		return
	case cp.EventQueue <- event:
	}

	//202 becasue async processing
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Node unregistered successfully"))

}

func (cp *APIServer) CreatePodHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close() // Close the request body to avoid resource leaks

	event := Event{
		ID:      time.Now().String(),
		action:  ActionCreatePod,
		Payload: "datAAAAAA",
	}

	//TODO WE NEED LOGIC TO SELECT WHICH NODE WILL TAKE THE POD AND THEN WE NEED TO BROADCAST THE OTHER PODS

	//try to push if we dont have space we discard until next (we also use queue to rate limit)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		http.Error(w, "Too many request", http.StatusTooManyRequests)
		return
	case cp.EventQueue <- event:
	}

	//202 becasue async processing
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Node unregistered successfully"))
}

func (cp *APIServer) WatchNodeHandler(w http.ResponseWriter, r *http.Request) {

	//TODO I WANT TO CREATE A STORE DB TO STORE KEY VALUE EVENTS ORDERED AND USE KEY TO KNOW WHICHS EVENT TO PROCESS
	//lastEventID, err := strconv.Atoi(r.Header.Get("Last-Event-ID")) // sent by the browser when it reconnects
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	responseController := http.NewResponseController(w)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "retry: 3000\n\n")
	if err := responseController.Flush(); err != nil {
		return
	}

	// a ping every 15s keeps proxies and NAT from dropping us (WE ALSO KNOW IF ITS ALIVE THE NODE)
	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer heartbeatTicker.Stop()

	defer log.Printf("%s disconnected", r.RemoteAddr)

	for {
		select {
		case <-r.Context().Done(): // the client went away
			return
		case event, stillOpen := <-cp.EventQueue:
			if !stillOpen {
				fmt.Fprint(w, "event: end\ndata: fin\n\n") // needs a data line, or EventSource drops the event
				responseController.Flush()
				return
			}
			// id, name and payload, one field per line; the blank line at the end closes the event
			fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, event.action, event.Payload)
			if err := responseController.Flush(); err != nil { // push it out of Go's buffer, now
				return
			}

		case <-heartbeatTicker.C:
			fmt.Fprint(w, ": ping\n\n")                        // a comment line: the client ignores it, the network sees traffic
			if err := responseController.Flush(); err != nil { // fails if the client is gone, so we stop
				return
			}
		}
	}
}

func (cp *APIServer) ListNodeDataHandler(w http.ResponseWriter, r *http.Request) {

}
