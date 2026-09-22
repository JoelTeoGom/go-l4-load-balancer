package network

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	// podNetworkPrefix is the cluster-wide pod network (10.244.0.0/16).
	// Every node owns one /24 inside it: node N -> 10.244.N.0/24.
	podNetworkPrefix = "10.244"
	podCIDRMaskBits  = 24
	bridgePrefix     = "br"

	minNodeNumber = 1
	maxNodeNumber = 254
)

// NodeNetwork holds the per-node network settings derived from the hostname.
type NodeNetwork struct {
	NodeNumber int    // 1
	PodCIDR    string // 10.244.1.0/24
	GatewayIP  string // 10.244.1.1/24 (assigned to the bridge)
	BridgeName string // br1
}

// NodeNetworkFromHostname derives the node network from the trailing number
// of the hostname (node-1 -> 10.244.1.0/24, gateway 10.244.1.1/24, bridge br1).
func NodeNetworkFromHostname(hostname string) (NodeNetwork, error) {
	nodeNumber, err := NodeNumberFromHostname(hostname)
	if err != nil {
		return NodeNetwork{}, err
	}

	return NodeNetwork{
		NodeNumber: nodeNumber,
		PodCIDR:    fmt.Sprintf("%s.%d.0/%d", podNetworkPrefix, nodeNumber, podCIDRMaskBits),
		GatewayIP:  fmt.Sprintf("%s.%d.1/%d", podNetworkPrefix, nodeNumber, podCIDRMaskBits),
		BridgeName: fmt.Sprintf("%s%d", bridgePrefix, nodeNumber),
	}, nil
}

// NodeNumberFromHostname returns the trailing digits of the hostname as an int
// (node-1 -> 1, node12 -> 12). It must fit in the third octet of the pod CIDR.
func NodeNumberFromHostname(hostname string) (int, error) {
	hostname = strings.TrimSpace(hostname)
	digitsStart := strings.LastIndexFunc(hostname, func(character rune) bool {
		return !unicode.IsDigit(character)
	}) + 1

	nodeNumber, err := strconv.Atoi(hostname[digitsStart:])
	if err != nil {
		return 0, fmt.Errorf("hostname %q must end with the node number (ex: node-1)", hostname)
	}
	if nodeNumber < minNodeNumber || nodeNumber > maxNodeNumber {
		return 0, fmt.Errorf("node number %d from hostname %q out of range [%d-%d]", nodeNumber, hostname, minNodeNumber, maxNodeNumber)
	}

	return nodeNumber, nil
}
