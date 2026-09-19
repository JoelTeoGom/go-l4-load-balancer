package main

import (
	"context"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/agent"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/config"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/controlplane"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/event"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/network"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/reconcile"
)

func main() {
	//TODO CREATE GRATEFUL SHUTDOWN
	ctx := context.Background()

	//configs
	cfg := config.NewConfig()

	//obtain Network variables
	var err error
	localIP, err := network.ObtainOutboundIP()
	hostname, err := network.ResolveHostname()
	if err != nil {
		panic(err)
	}

	eventJobs := make(chan event.Event)
	defer close(eventJobs)

	//Create Node and settup ip table rules
	node, err := agent.NewAgent(hostname, localIP, "10.244.1.0/24", "br0", "10.244.1.1/24", cfg.LbAddress())
	if err != nil {
		panic(err)
	}
	node.InitNodeSetup()

	//Register to ControlPlane
	cp := controlplane.NewControlPlane(node, eventJobs)
	err = cp.RegisterNode(ctx)
	if err != nil {
		panic(err)
	}

	//Create Agent Reconciler to Listen ControlPlane Events and Process (ex: Create pod, service, health...)
	go reconcile.StartReconciler(ctx, node, eventJobs)
	cp.Watch(ctx)

}
