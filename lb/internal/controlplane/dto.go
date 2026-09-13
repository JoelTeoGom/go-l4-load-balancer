package controlplane

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/JoelTeoGom/go-l4-load-balancer/lb/internal/registry"
)

// HealthResponse is the payload a node sends to report its health.
type HealthResponse struct {
	Status string `json:"status"`
}

// parseHealthResponse decodes a health payload and maps its status to the domain Status.
func parseHealthResponse(body io.Reader) (registry.Status, error) {
	var healthResponse HealthResponse
	if err := json.NewDecoder(body).Decode(&healthResponse); err != nil {
		return "", fmt.Errorf("decode health response: %w", err)
	}

	switch status := registry.Status(healthResponse.Status); status {
	case registry.StatusActive, registry.StatusInactive:
		return status, nil
	default:
		return "", fmt.Errorf("unknown health status %q", healthResponse.Status)
	}
}
