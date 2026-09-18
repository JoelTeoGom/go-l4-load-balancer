package node

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Pod struct {
	ID     string
	NodeID string
	IP     string // 10.244.1.11
	Port   string // 8080
	Status Status // Pending / Running / Failed
}

type Status string

const (
	StatusPending Status = "PENDING"
	StatusRunning Status = "RUNNING"
	StatusFailed  Status = "FAILED"
)

func (n *Node) CreatePod(serviceName string) (*Pod, error) {
	//COMO CREAR POD
	// ip netns add pod1
	// veth pair, un extremo al netns, el otro al bridge
	// IP dentro + lo up + ruta default al bridge
	// lanzar el proceso con ip netns exec
	// en el nodo: ip_forward=1 + MASQUERADE (salida) y DNAT (entrada)

	//ip netns add pod-x
	//add veth with iP and turn up,
	//peer that veth with bridge (entiendo que hace de switch) i redirigir todo el trafico hacia el bridge
	//quizas en el bridge voy a tener que crear una regla por cada pod, para rootearlos
	//lanzar el proceso entiendo con el codigo dentro

	service, ok := n.Services[serviceName]
	if !ok || service == nil {
		return nil, fmt.Errorf("Service unavailable to create a pod!")
	}

	podSlice := service.Pods
	podPrefix := len(service.Pods)
	id := fmt.Sprintf("%s-SERVICE-%s-POD-%d", n.ID, serviceName, podPrefix)

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
	pod := &Pod{
		ID:     id,
		NodeID: n.ID,
		Port:   service.Port,
		Status: StatusPending,
		IP:     podIP,
	}

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

func (n *Node) RemovePod(pod *Pod) (*Pod, error) {
	//ENVIAR UN SIGTERM I LUEGO SIGKILL I MATAMOS EL PROCESO
	return nil, nil
}

func (n *Node) CheckPodHealth(pod *Pod) (Status, error) {
	//PATCH contra pod para ver current STATUS
	return "", nil
}
