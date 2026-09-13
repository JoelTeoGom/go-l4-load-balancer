package registry

type Node struct {
	id      int
	address string

	status     Status
	cpuPercent int
}

func NewNode(id int, address string, status Status, cpuPercent int) *Node {
	return &Node{
		id:         id,
		address:    address,
		status:     status,
		cpuPercent: cpuPercent,
	}
}

func (n *Node) ID() int {
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
