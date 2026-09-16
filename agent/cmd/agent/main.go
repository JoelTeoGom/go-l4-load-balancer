package main

import (
	"context"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/config"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/controlplane"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/event"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/network"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/node"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/worker"
)

func main() {
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
	node := node.NewNode(hostname, localIP, " 10.244.1.0/24", cfg.LbAddress())
	node.InitSetup()

	//Register to ControlPlane
	cp := controlplane.NewControlPlane(node, eventJobs)
	err = cp.RegisterNode(ctx)
	if err != nil {
		panic(err)
	}

	//Create Agent Worker to Listen ControlPlane Events and Process (ex: Create pod, service, health...)
	worker.StartWorker(ctx, node, eventJobs)
}
