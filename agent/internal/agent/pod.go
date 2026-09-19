package agent

import (
	"fmt"
	"hash/fnv"
	"net"
	"strings"
	"time"
)

const (
	KubeServiceChainPrefix = "KUBE-SVC"
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

func (a *Agent) CreatePod(serviceName string) (*Pod, error) {
	service, ok := a.Services[serviceName]
	if !ok || service == nil {
		return nil, fmt.Errorf("Service unavailable to create a pod!")
	}

	podSlice := service.LocalPods

	podIP, err := a.GetNextIP()
	if err != nil {
		return nil, err
	}

	podPrefix := len(podSlice)

	kubeService := fmt.Sprintf("%s-%s", KubeServiceChainPrefix, serviceName)
	podID := fmt.Sprintf("%s-%d", kubeService, podPrefix)

	if err := SetupPodIptables(kubeService, podID, podIP, service.PodPort); err != nil {
		return nil, err
	}

	pod := NewPod(podID, serviceName, podIP)

	if err := a.SetupPodNetwork(pod); err != nil {
		return nil, err
	}

	podSlice = append(podSlice, pod)
	// 3. Cambiar los backends (crear o destruir un pod)

	// Se vacía la cadena y se reescribe entera, porque las probabilidades se recalculan.

	// Un backend:

	// iptables -t nat -F KUBE-SVC-API x
	// iptables -t nat -N KUBE-SEP-API-1 x
	// iptables -t nat -A KUBE-SVC-API -j KUBE-SEP-API-1 x
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

	//we give CTRL plane ACTION TO
	// action :=
	// event := newEvent(action, )
	// a.Broadcast()
	return nil, nil
}

// SetupPodIptables flushes the service chain and adds the pod endpoint chain with its DNAT rule
func SetupPodIptables(kubeService, podID, podIP, podPort string) error {
	args := fmt.Sprintf("iptables -t nat -F %s", kubeService)
	if err := run(args); err != nil {
		return err
	}

	args = fmt.Sprintf("iptables -t nat -N %s", podID)
	if err := run(args); err != nil {
		return err
	}

	args = fmt.Sprintf("iptables -t nat -A %s -j %s", kubeService, podID)
	if err := run(args); err != nil {
		return err
	}

	args = fmt.Sprintf("iptables -t nat -A %s -p tcp -j DNAT --to-destination %s:%s", podID, podIP, podPort)
	if err := run(args); err != nil {
		return err
	}

	return nil
}

// SetupPodNetwork creates the pod netns, wires it to the node bridge with a veth pair and configures IP + default route
func (a *Agent) SetupPodNetwork(pod *Pod) error {
	_, podNetwork, err := net.ParseCIDR(a.PodCIDR)
	if err != nil {
		return fmt.Errorf("invalid pod CIDR %q: %w", a.PodCIDR, err)
	}
	maskOnes, _ := podNetwork.Mask.Size()
	gatewayIP := strings.Split(a.Bridge.GatewayIP, "/")[0]

	vethHost, vethPod := vethNames(pod.ID)

	//2. Creating pod network namespace (same name as pod ID)
	args := fmt.Sprintf("ip netns add %s", pod.NetNS)
	if err := run(args); err != nil {
		return err
	}

	//3. Creating veth pair (one cable, two ends)
	args = fmt.Sprintf("ip link add %s type veth peer name %s", vethHost, vethPod)
	if err := run(args); err != nil {
		return err
	}

	//4. Moving pod end into the netns and renaming it to eth0
	args = fmt.Sprintf("ip link set %s netns %s", vethPod, pod.NetNS)
	if err := run(args); err != nil {
		return err
	}
	args = fmt.Sprintf("ip netns exec %s ip link set %s name eth0", pod.NetNS, vethPod)
	if err := run(args); err != nil {
		return err
	}

	//5-6. Plugging host end into the bridge
	args = fmt.Sprintf("ip link set %s master %s", vethHost, a.Bridge.Name)
	if err := run(args); err != nil {
		return err
	}
	args = fmt.Sprintf("ip link set %s up", vethHost)
	if err := run(args); err != nil {
		return err
	}

	//7. Configuring network inside the netns
	args = fmt.Sprintf("ip netns exec %s ip addr add %s/%d dev eth0", pod.NetNS, pod.IP, maskOnes)
	if err := run(args); err != nil {
		return err
	}
	args = fmt.Sprintf("ip netns exec %s ip link set eth0 up", pod.NetNS)
	if err := run(args); err != nil {
		return err
	}
	args = fmt.Sprintf("ip netns exec %s ip link set lo up", pod.NetNS)
	if err := run(args); err != nil {
		return err
	}
	args = fmt.Sprintf("ip netns exec %s ip route add default via %s", pod.NetNS, gatewayIP)
	if err := run(args); err != nil {
		return err
	}

	pod.VethHost = vethHost
	return nil
}

// vethNames returns short interface names derived from the pod ID
// Linux interface names are limited to 15 chars, so "veth" + podID would overflow
func vethNames(podID string) (string, string) {
	hasher := fnv.New32a()
	hasher.Write([]byte(podID))
	podHash := fmt.Sprintf("%08x", hasher.Sum32())
	return "veth" + podHash, "eth0" + podHash
}

func (n *Agent) RemovePod(pod *Pod) (*Pod, error) {
	//ENVIAR UN SIGTERM I LUEGO SIGKILL I MATAMOS EL PROCESO
	return nil, nil
}

func (n *Agent) CheckPodHealth(pod *Pod) (Status, error) {
	//PATCH contra pod para ver current STATUS
	return "", nil
}
