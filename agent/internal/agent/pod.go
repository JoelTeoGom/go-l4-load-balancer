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

// TODO(rollback): partial failures leave orphan state behind.
// The pod is only appended to LocalPods once everything succeeded, so the in-memory
// state stays clean, but the kernel state does not:
//   - GetNextIP marks the IP as allocated; no error path calls ReleaseIP.
//   - SetupPodNetwork may fail halfway, leaving the netns and the veth pair behind.
//   - SetupPodIptables flushes the service chain first, so a failure after that point
//     leaves it empty: no REJECT rule and no backends, i.e. the service blackholes.
//
// Each step needs its undo, run in reverse order on error.
func (a *Agent) CreatePod(serviceName string) (pod *Pod, err error) {
	service, ok := a.Services[serviceName]
	if !ok || service == nil {
		return nil, fmt.Errorf("Service unavailable to create a pod!")
	}
	podSlice := service.LocalPods

	podIP, err := a.GetNextIP()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			a.ReleaseIP(podIP)
		}
	}()

	podPrefix := len(podSlice)

	kubeService := fmt.Sprintf("%s-%s", KubeServiceChainPrefix, serviceName)
	podID := fmt.Sprintf("%s-%d", kubeService, podPrefix)

	pod = NewPod(podID, serviceName, podIP)

	if err = a.SetupPodNetwork(pod); err != nil {
		return nil, err
	}

	if err = service.SetupPodIptables(kubeService, pod); err != nil {
		return nil, err
	}
	service.LocalPods = append(service.LocalPods, pod)
	// 3. Cambiar los backends (crear o destruir un pod)
	// Tres backends, el último en otro nodo:

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

// SetupPodIptables creates the new pod endpoint chain and rewrites the service chain with a jump to every backend
// Probabilities are 1/n, 1/(n-1), ..., and the last rule has no probability so it takes whatever is left
func (s *Service) SetupPodIptables(kubeService string, pod *Pod) error {

	//1. Flushing service chain, it is rewritten entirely because probabilities change
	args := fmt.Sprintf("iptables -t nat -F %s", kubeService)
	if err := run(args); err != nil {
		return err
	}

	//2. Creating Backends
	backends := s.Backends()
	backends = append(backends, Backend{IP: pod.IP, Port: s.PodPort})
	backendCount := len(backends)
	//3. Creating new pod endpoint chain with its DNAT to the pod
	args = fmt.Sprintf("iptables -t nat -N %s", pod.ID)
	if err := run(args); err != nil {
		return err
	}
	args = fmt.Sprintf("iptables -t nat -A %s -p tcp -j DNAT --to-destination %s:%s", pod.ID, pod.IP, s.PodPort)
	if err := run(args); err != nil {
		return err
	}

	//4. Jumping from service chain to every backend
	for backendIndex, backend := range backends {
		if backendIndex == backendCount-1 {
			args = fmt.Sprintf("iptables -t nat -A %s -p tcp -j %s", kubeService, backend.ID)
		} else {
			probability := 1.0 / float64(backendCount-backendIndex)
			args = fmt.Sprintf("iptables -t nat -A %s -p tcp -m statistic --mode random --probability %.5f -j %s", kubeService, probability, backend.ID)
		}
		if err := run(args); err != nil {
			return err
		}
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
