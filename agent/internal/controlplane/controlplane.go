package controlplane

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/event"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/node"
)

// ControlPlane is the HTTP client the agent uses to talk to the load balancer's control plane.
type ControlPlane struct {
	Client *http.Client
	Node   *node.Node
}

func NewControlPlane(node *node.Node) *ControlPlane {
	return &ControlPlane{
		Node: node,
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

func (cp *ControlPlane) Watch(ctx context.Context, eventJob chan event.Event) {

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
