package controlplane

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/event"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/node"
)

// ControlPlane is the HTTP client the agent uses to talk to the load balancer's control plane.
type ControlPlane struct {
	Client     *http.Client
	Node       *node.Node
	EventQueue chan<- event.Event
}

func NewControlPlane(node *node.Node, eventQueue chan<- event.Event) *ControlPlane {
	return &ControlPlane{
		EventQueue: eventQueue,
		Node:       node,
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (cp *ControlPlane) RegisterNode(ctx context.Context) error {
	return cp.post(ctx, "/register-node", cp.Node.ID)
}

func (cp *ControlPlane) UnregisterNode(ctx context.Context) error {
	return cp.post(ctx, "/unregister-node", cp.Node.ID)
}

func (cp *ControlPlane) Watch(ctx context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.StreamClient.url+"/watch-node", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if w.StreamClient.lastID != "" {
		req.Header.Set("Last-Event-ID", w.StreamClient.lastID) // TODO server can start FROM LAST ID STORED
	}
	resp, err := w.StreamClient.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error connecting watchdog to ctrlPlane %s", w.Node.ID)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		return fmt.Errorf("unexpected content-type %q", ct)
	}

	sc := bufio.NewScanner(resp.Body)

	// El buffer por defecto es 64 KB por línea: un JSON grande en un data: lo rompe.
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	//TODO PROGRAM WATCH CLIENT LOGIC
	for {

		//encolar con cancelacion
		select {}
	}
}

func (cp *ControlPlane) post(ctx context.Context, path string, payload any) error {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	endpoint := cp.Node.LoadBalancerUrl + path
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := cp.Client.Do(request)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, response.Body) // drain so the conn can be reused
		response.Body.Close()
	}()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		errorBody, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return fmt.Errorf("unexpected status %d: %s", response.StatusCode, errorBody)
	}
	return nil
}
