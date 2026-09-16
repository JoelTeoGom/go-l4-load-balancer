package node

import (
	"fmt"
	"os/exec"
	"strings"
)

type Service struct {
	Name string
	IP   string
	Port string
	Pods []*Pod
}

func (n *Node) CreateService(serviceName, IP, port string) (*Service, error) {
	if service, ok := n.Services[serviceName]; ok {
		return service, nil
	}

	newService := &Service{
		Name: serviceName,
		IP:   IP,
		Port: port,
		Pods: make([]*Pod, 0),
	}
	n.Services[serviceName] = newService

	rule := fmt.Sprintf("-A KUBE-SERVICES -d %s -p tcp --dport %s -j KUBE-SVC", IP, port)
	args := strings.Fields(rule)
	cmd := exec.Command("iptables", args...)
	cmd.CombinedOutput()

	return newService, nil
}
