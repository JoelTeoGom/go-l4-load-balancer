package agent

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Pod struct {
	ID          string
	ServiceName string
	IP          string // 10.244.1.5
	NetNS       string // NetNS name
	VethHost    string // veth host peer
	Status      Status // pending | running | failed
	CreatedAt   time.Time
}
type Status string

const (
	StatusPending Status = "PENDING"
	StatusRunning Status = "RUNNING"
	StatusFailed  Status = "FAILED"
)

func NewPod(id, serviceName, ip string) *Pod {
	return &Pod{
		ID:          id,
		ServiceName: serviceName,
		IP:          ip,
		NetNS:       id,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
	}
}

func (n *Agent) CreatePod(serviceName string) (*Pod, error) {
	// 3. Cambiar los backends (crear o destruir un pod)

	// Se vacía la cadena y se reescribe entera, porque las probabilidades se recalculan.

	// Un backend:

	// iptables -t nat -F KUBE-SVC-API
	// iptables -t nat -N KUBE-SEP-API-1
	// iptables -t nat -A KUBE-SVC-API -j KUBE-SEP-API-1
	// iptables -t nat -A KUBE-SEP-API-1 -p tcp -j DNAT --to-destination 10.244.1.5:8080

	// Dos backends:

	// iptables -t nat -F KUBE-SVC-API
	// iptables -t nat -N KUBE-SEP-API-2
	// iptables -t nat -A KUBE-SVC-API -m statistic --mode random --probability 0.50000 -j KUBE-SEP-API-1
	// iptables -t nat -A KUBE-SVC-API -j KUBE-SEP-API-2
	// iptables -t nat -A KUBE-SEP-API-2 -p tcp -j DNAT --to-destination 10.244.1.6:8080

	// Tres, el último en otro nodo:

	// iptables -t nat -F KUBE-SVC-API
	// iptables -t nat -N KUBE-SEP-API-C
	// iptables -t nat -A KUBE-SVC-API -m statistic --mode random --probability 0.33333 -j KUBE-SEP-API-1
	// iptables -t nat -A KUBE-SVC-API -m statistic --mode random --probability 0.50000 -j KUBE-SEP-API-2
	// iptables -t nat -A KUBE-SVC-API -j KUBE-SEP-API-C
	// iptables -t nat -A KUBE-SEP-API-C -p tcp -j DNAT --to-destination 192.168.1.22:30081

	// Probabilidades 1/n, 1/(n-1), …, y la última regla sin --probability.
	service, ok := n.Services[serviceName]
	if !ok || service == nil {
		return nil, fmt.Errorf("Service unavailable to create a pod!")
	}

	podSlice := service.Pods
	podPrefix := len(service.Pods)
	id := fmt.Sprintf("%s-SERVICE-%s-POD-%d", n.NodeID, serviceName, podPrefix)

	ipMASK := strings.Split(n.PodCIDR, "/")
	if len(ipMASK) < 2 {
		return nil, fmt.Errorf("No ip format specified")
	}
	mask, err := strconv.Atoi(ipMASK[1])
	if err != nil {
		return nil, fmt.Errorf("Wrong mask format")
	}

	//VALID IP range 0 - 255
	maxRange := 1<<mask - 1    // (1<<8) - 1 = 255
	hostValue := len(podSlice) //we are going to give ordered IPs
	if hostValue > maxRange {
		return nil, fmt.Errorf("Max ips reached")
	}

	ip := ipMASK[0]

	ipBytes := []byte(ip)
	ipBytes = ipBytes[:len(ipBytes)-1]
	podIP := fmt.Sprintf("%s%d", string(ipBytes), hostValue)
	pod := NewPod(id, serviceName, podIP)

	//1. Create netns pod x
	rule := fmt.Sprintf("add %s", pod.ID)
	args := strings.Fields(rule)
	cmd := exec.Command("ipnetns", args...)
	cmd.CombinedOutput()

	//2.

	// podSlice := service.Pods
	// pod :=
	// rule := "-t nat -A PREROUTING -j KUBE-SERVICES"
	// args := strings.Fields(rule) // ["-t","nat","-A","PREROUTING","-j","KUBE-SERVICES"]
	// cmd = exec.Command("iptables", args...)
	// cmd.CombinedOutput()

	//add if everything worked fine
	service.Pods = append(service.Pods, pod)
	return nil, nil

}

// ip netns add pod1
// veth pair, un extremo al netns, el otro al bridge
// IP dentro + lo up + ruta default al bridge
// lanzar el proceso con ip netns exec
// en el nodo: ip_forward=1 + MASQUERADE (salida) y DNAT (entrada)

func (n *Agent) RemovePod(pod *Pod) (*Pod, error) {
	//ENVIAR UN SIGTERM I LUEGO SIGKILL I MATAMOS EL PROCESO
	return nil, nil
}

func (n *Agent) CheckPodHealth(pod *Pod) (Status, error) {
	//PATCH contra pod para ver current STATUS
	return "", nil
}
