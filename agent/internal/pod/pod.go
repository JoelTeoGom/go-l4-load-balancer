package pod

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Node struct {
	ID       string // ej: NODE-1
	IP       string // ej: 192.168.1.51
	PodCIDR  string // ej: 10.244.1.0/24
	LastSeen time.Time
	services map[string]Service
}

type Service struct {
	Name     string
	NodePort int // 30080
	Selector string
	Pods     []*Pod
}

type Pod struct {
	ID     string
	NodeID string
	IP     string // 10.244.1.11
	Port   int    // 8080
	Status Status // Pending / Running / Failed
}

type Status string

const (
	StatusPending Status = "PENDING"
	StatusRunning Status = "RUNNING"
	StatusFailed  Status = "FAILED"
)

func (n *Node) CreatePod(serviceName string) error {
	service, ok := n.services[serviceName]
	if !ok {
		return fmt.Errorf("Service unavailable!")
	}

	podSlice := service.Pods

	pod := 



	rule := "-t nat -A PREROUTING -j KUBE-SERVICES"
	args := strings.Fields(rule) // ["-t","nat","-A","PREROUTING","-j","KUBE-SERVICES"]
	cmd = exec.Command("iptables", args...)
	cmd.CombinedOutput()

}

// ip netns add pod1
// veth pair, un extremo al netns, el otro al bridge
// IP dentro + lo up + ruta default al bridge
// lanzar el proceso con ip netns exec
// en el nodo: ip_forward=1 + MASQUERADE (salida) y DNAT (entrada)
