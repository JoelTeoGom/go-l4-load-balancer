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

	newService.CreateServiceSetup()

	return newService, nil
}

func (n *Service) CreateServiceSetup() {
	//1. Creating service chain
	run("iptables -t nat -N KUBE-SVC-API")

	//2. Jumping from KUBE-SERVICES to service chain (ClusterIP and NodePort)
	args := fmt.Sprintf("iptables -t nat -A KUBE-SERVICES -d %s -p tcp --dport %s -j KUBE-SVC-API", n.ClusterIP, n.ClusterPort)
	run(args)

	for _, remoteNodeIP := range n.RemoteNode {
		args = fmt.Sprintf("iptables -t nat -A KUBE-SERVICES -d %s -p tcp --dport %s -j KUBE-SVC-API", remoteNodeIP, n.NodePort)
		run(args)
	}

	//3. Rejecting traffic while service has no pods
	run("iptables -t nat -A KUBE-SVC-API -j REJECT --reject-with icmp-port-unreachable")
}
