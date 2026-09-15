package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/event"
)

type Node struct {
	ID              string // ej: NODE-1
	IP              string // ej: 192.168.1.51
	PodCIDR         string // ej: 10.244.1.0/24
	LastSeen        time.Time
	LoadBalancerUrl string
	Services        map[string]Service
}

func NewNode(hostname, ipAddr, podCIDR, url string) *Node {
	return &Node{
		ID:              hostname,
		IP:              ipAddr,
		PodCIDR:         podCIDR,
		LoadBalancerUrl: url,
		LastSeen:        time.Now(),
		Services:        make(map[string]Service),
	}
}

func (n *Node) InitSetup() error {
	//1. Setting up TCP forwarding to 1
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sysctl: %w: %s", err, out)
	}

	//2. Setting up PREROUTING jump TO KUBE-SERVICES
	rule := "-t nat -A PREROUTING -j KUBE-SERVICES"
	args := strings.Fields(rule)
	cmd = exec.Command("iptables", args...)
	cmd.CombinedOutput()

	return nil
}

func (n *Node) RegisterNode(ctx context.Context, addr, hostname string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	body, err := json.Marshal(hostname)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	endpoint := fmt.Sprintf("%s/register-node", n.LoadBalancerUrl)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
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

func (n *Node) WatchControlPlane(ctx context.Context, eventJob chan event.Event) {

}
