package registry

type Node struct {
	id         string
	address    string
	status     Status
	cpuPercent int
}

func NewNode(id string, address string, status Status, cpuPercent int) *Node {
	return &Node{
		id:         id,
		address:    address,
		status:     status,
		cpuPercent: cpuPercent,
	}
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) Address() string {
	return n.address
}

func (n *Node) Status() Status {
	return n.status
}

func (n *Node) CPUPercent() int {
	return n.cpuPercent
}
