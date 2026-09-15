package node

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

	// service, ok := n.services[serviceName]
	// if !ok {
	// 	return fmt.Errorf("Service unavailable!")
	// }

	// podSlice := service.Pods
	// pod :=
	// rule := "-t nat -A PREROUTING -j KUBE-SERVICES"
	// args := strings.Fields(rule) // ["-t","nat","-A","PREROUTING","-j","KUBE-SERVICES"]
	// cmd = exec.Command("iptables", args...)
	// cmd.CombinedOutput()

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
