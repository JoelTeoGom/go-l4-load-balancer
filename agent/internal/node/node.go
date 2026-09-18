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

func (n *Node) InitNodeSetup() error {
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
	run("ip link add name br0 type bridge")
	run("ip addr add 10.244.1.1/24 dev br0")
	run("ip link set br0 up")

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
	run("iptables -t nat -A KUBE-POSTROUTING -s 10.244.1.0/24 ! -d 10.244.1.0/24 -j MASQUERADE")

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
