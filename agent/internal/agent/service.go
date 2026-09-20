package agent

import "fmt"

type Service struct {
	Name string

	ClusterIP   string
	ClusterPort string //the une used combined with Cluster IP

	PodPort string //the one that we use inside PODs (maybe we need it to curl)  HEALTH ISSUES

	NodePort string //The one combined with NodeIP:NodePort used to identify remote SERVICE

	LocalPods  []*Pod
	RemoteNode []Backend //NODES that have available pods (THIS service pods)

}
type Backend struct {
	ID   string
	IP   string
	Port string
}

func NewService(name, clusterIP, clusterPort, nodePort, podPort string) *Service {
	return &Service{
		Name:        name,
		ClusterIP:   clusterIP,
		ClusterPort: clusterPort,
		PodPort:     podPort,
		NodePort:    nodePort,
		LocalPods:   []*Pod{},
		RemoteNode:  []Backend{},
	}
}

func (n *Agent) CreateService(name, clusterIP, clusterPort, nodePort, podPort string) (*Service, error) {
	if service, ok := n.Services[name]; ok {
		return service, nil
	}

	newService := NewService(name, clusterIP, clusterPort, nodePort, podPort)
	n.Services[name] = newService

	newService.CreateServiceSetup(name)

	return newService, nil
}

func (s *Service) CreateServiceSetup(name string) {
	//1. Creating service chain

	serviceName := fmt.Sprintf("%s-%s", KubeServiceChainPrefix, name)
	args := fmt.Sprintf("iptables -t nat -N %s", serviceName)
	run(args)

	//2. Jumping from KUBE-SERVICES to service chain (ClusterIP and NodePort)
	args = fmt.Sprintf("iptables -t nat -A KUBE-SERVICES -d %s -p tcp --dport %s -j %s", s.ClusterIP, s.ClusterPort)
	run(args)

	for _, backend := range s.RemoteNode {
		args = fmt.Sprintf("iptables -t nat -A KUBE-SERVICES -d %s -p tcp --dport %s -j %s", backend.IP, s.NodePort)
		run(args)
	}

	//3. Rejecting traffic while service has no pods
	run("iptables -t nat -A KUBE-SVC-API -j REJECT --reject-with icmp-port-unreachable")
}

// Backends returns local pods first and then remote nodes that also serve this service
func (s *Service) Backends() []Backend {
	var backends []Backend
	for _, pod := range s.LocalPods {
		backends = append(backends, Backend{ID: pod.ID, IP: pod.IP, Port: s.PodPort})
	}
	for _, externalBackend := range s.RemoteNode {
		backends = append(backends, externalBackend)
	}
	return backends
}
