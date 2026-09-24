package registry

import (
	"context"
	"math/rand/v2"
	"sync"
)

type Registry struct {
	mu     sync.Mutex
	nodes  []*Node
	events chan Event
}

func NewRegistry() *Registry {
	return &Registry{
		nodes:  []*Node{},
		events: make(chan Event, 10),
	}
}

func (r *Registry) GetNodeByAddress(address string) *Node {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, node := range r.nodes {
		if node.Address() == address {
			return node
		}
	}
	return nil
}

func (r *Registry) EmitEvent(ctx context.Context, node *Node, event string) {
	select {
	case <-ctx.Done():
		return
	case r.events <- Event{
		Node:  node,
		Event: event,
	}:
	default:
		// to avoid blocking if the channel is full, we will get another one after a while
	}
}

func (r *Registry) AddNode(node *Node) *Node {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes = append(r.nodes, node)
	return node
}

func (r *Registry) RemoveNode(node *Node) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, n := range r.nodes {
		if n == node {
			r.nodes = append(r.nodes[:i], r.nodes[i+1:]...)
			break
		}
	}
}

func (r *Registry) ListNodes() []*Node {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.nodes
}

// Test function to obtain a random node from the registry
func (r *Registry) ObtainRandomNode() *Node {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.nodes) == 0 {
		return nil
	}
	index := randomNodeIndex(len(r.nodes))
	return r.nodes[index]
}

func randomNodeIndex(nodeCount int) int {
	return rand.IntN(nodeCount)
}
