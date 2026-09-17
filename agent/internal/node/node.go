package node

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Node struct {
	ID              string // ej: NODE-1
	IP              string // ej: 192.168.1.51
	PodCIDR         string // ej: 10.244.1.0/24
	LastSeen        time.Time
	LoadBalancerUrl string
	Services        map[string]*Service

	//0-255 IPs available between services in 1 NODE

}

func NewNode(hostname, ipAddr, podCIDR, url string) *Node {
	return &Node{
		ID:              hostname,
		IP:              ipAddr,
		PodCIDR:         podCIDR,
		LoadBalancerUrl: url,
		LastSeen:        time.Now(),
		Services:        make(map[string]*Service),
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

	//3. Setting up Destination NAT
	rule = "-t nat -A POSTROUTING -o eth0 -j MASQUERADE"
	args = strings.Fields(rule)
	cmd = exec.Command("iptables", args...)
	cmd.CombinedOutput()

	//4. Setting up bridge
	// crear el bridge
	exec.Command("ip", "link", "add", "cni0", "type", "bridge")
	exec.Command("ip", "addr", "add", "10.244.1.1/24", "dev", "cni0")
	exec.Command("ip", "link", "set", "cni0", "up")

	return nil
}
