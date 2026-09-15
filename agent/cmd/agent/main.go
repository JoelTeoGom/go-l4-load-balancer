package main

import (
	"context"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/config"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/event"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/network"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/node"
)

func main() {
	ctx := context.Background()

	cfg := config.NewConfig()

	//obtain Network variables
	var err error
	localIP, err := network.ObtainOutboundIP()
	hostname, err := network.ResolveHostname()
	if err != nil {
		panic(err)
	}

	//Create Node and Start Node
	eventJobs := make(chan event.Event)
	node := node.NewNode(hostname, localIP, " 10.244.1.0/24", cfg.LbAddress())
	node.InitSetup()
	node.StartServer(ctx, localIP, cfg.AgentPort(), eventJobs)
	//	hostName, err := initSetup(cfg.LbAddress(), cfg.SvcPort())

	if err != nil {
		panic(err)
	}

	// controlPlane := controlplane.NewControlPlane(node)
	// err = controlPlane.RegisterNode(ctx)
	// if err != nil {
	// 	panic(err)
	// }

}
