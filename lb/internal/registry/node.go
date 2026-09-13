package registry

import "net"

type Node struct {
	id         int
	conn       net.Conn
	status     Status
	cpuPercent float64
}

func (n *Node) ID() int {
	return n.id
}

func (n *Node) Conn() net.Conn {
	return n.conn
}

func NewNode(id int, conn net.Conn) *Node {
	return &Node{
		id:   id,
		conn: conn,
	}
}
