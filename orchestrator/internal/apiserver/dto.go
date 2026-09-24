package apiserver

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/JoelTeoGom/kubernetes-from-scratch/orchestrator/internal/registry"
)

type HealthResponse struct {
	Status string `json:"status"`
}

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

type RegisterNodeRequest struct {
	NodeID string
	NodeIP string
}
