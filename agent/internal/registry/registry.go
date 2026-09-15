package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Registry struct {
	cp       *ControlPlane
	hostname string
}

type ControlPlane struct {
	address string
	client  *http.Client
}

func NewRegistry(address, hostname string) *Registry {
	return &Registry{
		hostname: hostname,
		cp: &ControlPlane{
			address: address,
			client: &http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 10,
					IdleConnTimeout:     90 * time.Second,
				},
			},
		},
	}
}

func (r *Registry) ConnectToCtrlPlane(ctx context.Context, addr, hostname string) error {
	body, err := json.Marshal(hostname)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	endpoint := fmt.Sprintf("http://%s/register-node")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := r.cp.client.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, res.Body) // drain so the conn can be reused
		res.Body.Close()
	}()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 4<<10))
		return fmt.Errorf("unexpected status %d: %s", res.StatusCode, b)
	}

	return nil
}
