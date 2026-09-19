package agent

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type Agent struct {
	NodeID          string // ej: NODE-1
	NodeIP          string // ej: 192.168.1.51
	PodCIDR         string // ej: 10.244.1.0/24
	Bridge          Bridge
	ControlPlaneURL string
	Services        map[string]*Service

	//IPs already allocated inside PodCIDR shared by all services in 1 NODE (gateway included)
	//len(AllocatedIPs) is used as pointer to the next free IP
	//(valid pod IPs for /24: .2 - .254 -> 253 IPs; .0 network, .1 gateway, .255 broadcast are reserved)
	AllocatedIPs []string
}

type Bridge struct {
	Name      string // br0
	GatewayIP string
}

func NewAgent(nodeID, nodeIP, podCIDR, bridgeName, gatewayIP, controlPlaneURL string) (*Agent, error) {
	_, podNetwork, err := net.ParseCIDR(podCIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid pod CIDR %q: %w", podCIDR, err)
	}

	//Allocatable IPs (.1 - .254 for a /24): gateway (.1) + pod IPs (.2 - .254), skipping network (.0) and broadcast (.255)
	maskOnes, maskBits := podNetwork.Mask.Size()
	allocatedIPsCapacity := 1<<(maskBits-maskOnes) - 2
	if allocatedIPsCapacity < 0 {
		allocatedIPsCapacity = 0
	}

	allocatedIPs := make([]string, 0, allocatedIPsCapacity)
	allocatedIPs = append(allocatedIPs, gatewayIP)

	return &Agent{
		NodeID:  nodeID,
		NodeIP:  nodeIP,
		PodCIDR: podCIDR,
		Bridge: Bridge{
			Name:      bridgeName,
			GatewayIP: gatewayIP,
		},
		ControlPlaneURL: controlPlaneURL,
		Services:        make(map[string]*Service),
		AllocatedIPs:    allocatedIPs,
	}, nil
}

func (a *Agent) InitNodeSetup() error {
	//1. Loading br_netfilter so bridged traffic goes through iptables
	if err := run("modprobe br_netfilter"); err != nil {
		return err
	}

	//2. Enabling IP forwarding and iptables on bridge
	if err := run("sysctl -w net.ipv4.ip_forward=1"); err != nil {
		return err
	}
	if err := run("sysctl -w net.bridge.bridge-nf-call-iptables=1"); err != nil {
		return err
	}

	//3. Setting up bridge
	args := fmt.Sprintf("ip link add name %s type bridge", a.Bridge.Name)
	run(args)
	args = fmt.Sprintf("ip addr add %s dev %s", a.Bridge.GatewayIP, a.Bridge.Name)
	run(args)
	args = fmt.Sprintf("ip link set %s up", a.Bridge.Name)
	run(args)

	//4. Creating KUBE chains
	run("iptables -t nat -N KUBE-SERVICES")
	run("iptables -t nat -N KUBE-POSTROUTING")
	run("iptables -t filter -N KUBE-FORWARD")

	//5. Jumping from built-in chains to KUBE chains
	run("iptables -t nat -I PREROUTING 1 -j KUBE-SERVICES")
	run("iptables -t nat -I OUTPUT 1 -j KUBE-SERVICES")
	run("iptables -t nat -I POSTROUTING 1 -j KUBE-POSTROUTING")
	run("iptables -t filter -I FORWARD 1 -j KUBE-FORWARD")

	//6. Masquerading pod traffic leaving the pod CIDR
	args = fmt.Sprintf("iptables -t nat -A KUBE-POSTROUTING -s %s ! -d %s -j MASQUERADE", a.PodCIDR, a.PodCIDR)
	run(args)

	return nil
}

func run(command string) error {
	args := strings.Fields(command)
	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w: %s", command, err, out)
	}
	return nil
}

// helper we use to return IPs inside a NODE private net
// Valid IPs (for PodCIDR 10.244.1.0/24): 10.244.1.2 - 10.244.1.254 (253 IPs)
// Reserved: 10.244.1.0 (network), 10.244.1.1 (gateway/bridge), 10.244.1.255 (broadcast)
func (a *Agent) GetNextIP() (string, error) {
	_, podNetwork, err := net.ParseCIDR(a.PodCIDR)
	if err != nil {
		return "", fmt.Errorf("invalid pod CIDR %q: %w", a.PodCIDR, err)
	}
	maskOnes, maskBits := podNetwork.Mask.Size()
	allocatedIPsCapacity := 1<<(maskBits-maskOnes) - 2
	if allocatedIPsCapacity < 0 {
		allocatedIPsCapacity = 0
	}

	nextIP := len(a.AllocatedIPs) + 1
	if nextIP >= allocatedIPsCapacity {
		return "", fmt.Errorf("Exceeded number of IPs")
	}

	basicIP := podNetwork.IP.String()
	ipBytes := []byte(basicIP)
	ipBytes = ipBytes[:len(ipBytes)-1]
	podIP := fmt.Sprintf("%s%d", string(ipBytes), nextIP)
	a.AllocatedIPs = append(a.AllocatedIPs, podIP)
	return podIP, nil
}
