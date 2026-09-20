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

	//pod IDs already handed out for this service, same idea as Agent.allocated for IPs
	//a released ID goes back to false so the index can be reused without shifting the others
	takenIDs map[string]bool
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
		takenIDs:    map[string]bool{},
	}
}

func (n *Agent) CreateService(name, clusterIP, clusterPort, nodePort, podPort string) (*Service, error) {
	if service, ok := n.Services[name]; ok {
		return service, nil
	}

	newService := NewService(name, clusterIP, clusterPort, nodePort, podPort)
	n.Services[name] = newService

	newService.CreateServiceSetup(name, n.NodeIP)

	return newService, nil
}

func (s *Service) CreateServiceSetup(name, nodeIP string) {
	//1. Creating service chain

	serviceName := fmt.Sprintf("%s-%s", KubeServiceChainPrefix, name)
	run("iptables", "-t", "nat", "-N", serviceName)

	//2. Jumping from KUBE-SERVICES to service chain (ClusterIP and NodePort)
	run("iptables", "-t", "nat", "-A", "KUBE-SERVICES", "-d", s.ClusterIP, "-p", "tcp", "--dport", s.ClusterPort, "-j", serviceName)

	run("iptables", "-t", "nat", "-A", "KUBE-SERVICES", "-d", nodeIP, "-p", "tcp", "--dport", s.NodePort, "-j", serviceName)

	//3. Rejecting traffic while service has no pods
	run("iptables", "-t", "nat", "-A", serviceName, "-j", "REJECT", "--reject-with", "icmp-port-unreachable")
}

// GetNextPodID returns the lowest free pod ID for this service and marks it as taken
// Scanning a map instead of using len(LocalPods) keeps IDs stable when a pod in the
// middle is removed, the same reason Agent.GetNextIP scans instead of counting
func (s *Service) GetNextPodID() (string, error) {
	// with n IDs known, index n is free at worst, so the scan always terminates
	for index := 0; index <= len(s.takenIDs); index++ {
		podID := fmt.Sprintf("%s-%s-%d", KubeServiceChainPrefix, s.Name, index)
		if !s.takenIDs[podID] {
			s.takenIDs[podID] = true
			return podID, nil
		}
	}
	return "", fmt.Errorf("no free pod ID for service %s", s.Name)
}

// ReleaseID gives a pod ID back so a later pod can reuse the index
func (s *Service) ReleaseID(releasedID string) error {
	if !s.takenIDs[releasedID] {
		return fmt.Errorf("pod ID %s already released", releasedID)
	}
	s.takenIDs[releasedID] = false
	return nil
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
