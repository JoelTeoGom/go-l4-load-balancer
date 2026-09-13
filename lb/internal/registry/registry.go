package registry

import (
	"math/rand/v2"
	"net"
)

type Registry struct {
	nodes []*Node
}

func NewRegistry() *Registry {
	return &Registry{
		nodes: []*Node{},
	}
}

type Node struct {
	id   int
	conn net.Conn
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

func (r *Registry) AddNode(node *Node) {
	r.nodes = append(r.nodes, node)
}

func (r *Registry) RemoveNode(node *Node) {
	for i, n := range r.nodes {
		if n == node {
			r.nodes = append(r.nodes[:i], r.nodes[i+1:]...)
			break
		}
	}
}

func (r *Registry) ListNodes() []*Node {
	return r.nodes
}

// Test function to obtain a random node from the registry
func (r *Registry) ObtainRandomNode() *Node {
	if len(r.nodes) == 0 {
		return nil
	}
	index := randomNodeIndex(len(r.nodes))
	return r.nodes[index]
}

func randomNodeIndex(nodeCount int) int {
	return rand.IntN(nodeCount)
}
