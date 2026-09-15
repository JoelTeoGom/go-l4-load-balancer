package controlplane

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

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

func (cp *ControlPlane) CreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close() // Close the request body to avoid resource leaks

	event := Event{
		ID:      time.Now().String(),
		action:  ActionCreateService,
		Payload: "Name",
	}

	cp.EventQueue

	w.WriteHeader(r.Response.StatusCode)
	w.Write([]byte("Node unregistered successfully"))

}

func (cp *ControlPlane) CreatePodHandler(w http.ResponseWriter, r *http.Request) {

}

func (cp *ControlPlane) WatchNodeHandler(w http.ResponseWriter, r *http.Request) {
	lastEventID, err := strconv.Atoi(r.Header.Get("Last-Event-ID")) // sent by the browser when it reconnects

	w.Header().Set("Content-Type", "text/event-stream") // the body is a stream of events, not a document
	w.Header().Set("Cache-Control", "no-cache")         // never store or replay this response
	// w.Header().Set("Connection", "keep-alive")       // optional no-op: default in HTTP/1.1, dropped in HTTP/2

	responseController := http.NewResponseController(w) // gives us Flush without asserting http.Flusher
	w.WriteHeader(http.StatusOK)                        // headers are ready; the body stays open
	fmt.Fprint(w, "retry: 3000\n\n")                    // if the connection drops, reconnect after 3s
	if err := responseController.Flush(); err != nil {  // send the headers now so the client sees the stream open
		return
	}

	events := make(chan Event)
	go reciteHamlet(r.Context(), startLine, events) // someone else produces events; the handler only writes them

	heartbeatTicker := time.NewTicker(15 * time.Second) // a ping every 15s keeps proxies and NAT from dropping us
	defer heartbeatTicker.Stop()

	log.Printf("%s connected, Last-Event-ID=%q, starting at line %d", r.RemoteAddr, r.Header.Get("Last-Event-ID"), startLine)
	defer log.Printf("%s disconnected", r.RemoteAddr)

	for {
		select {
		case <-r.Context().Done(): // the client went away
			return

		case event, stillOpen := <-events:
			if !stillOpen { // the producer closed the channel: nothing left to recite
				fmt.Fprint(w, "event: end\ndata: fin\n\n") // needs a data line, or EventSource drops the event
				responseController.Flush()
				return
			}
			// id, name and payload, one field per line; the blank line at the end closes the event
			fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, event.Name, event.Data)
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
