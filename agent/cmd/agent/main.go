package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/JoelTeoGom/kubernetes-from-scratch/agent/internal/agent"
	"github.com/JoelTeoGom/kubernetes-from-scratch/agent/internal/apiclient"
	"github.com/JoelTeoGom/kubernetes-from-scratch/agent/internal/config"
	"github.com/JoelTeoGom/kubernetes-from-scratch/agent/internal/event"
	"github.com/JoelTeoGom/kubernetes-from-scratch/agent/internal/network"
	"github.com/JoelTeoGom/kubernetes-from-scratch/agent/internal/reconcile"
)

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//configs
	cfg := config.NewConfig()

	//obtain Network variables
	localIP, err := network.ObtainOutboundIP()
	if err != nil {
		panic(err)
	}
	hostname, err := network.ResolveHostname()
	if err != nil {
		panic(err)
	}
	//PodCIDR, gateway and bridge are derived from the node number at the end of the hostname (node-1)
	nodeNetwork, err := network.NodeNetworkFromHostname(hostname)
	if err != nil {
		panic(err)
	}

	eventJobs := make(chan event.Event)
	defer close(eventJobs)

	//Create Node and settup ip table rules
	node, err := agent.NewAgent(hostname, localIP, nodeNetwork.PodCIDR, nodeNetwork.BridgeName, nodeNetwork.GatewayIP, cfg.ControlPlaneURL())
	if err != nil {
		panic(err)
	}
	if err := node.InitNodeSetup(); err != nil {
		panic(err)
	}

	//Register to ControlPlane
	cp := apiclient.NewAPIClient(node, eventJobs)
	err = cp.RegisterNode(ctx)
	if err != nil {
		panic(err)
	}

	//Create Agent Reconciler to Listen ControlPlane Events and Process (ex: Create pod, service, health...)
	go reconcile.StartReconciler(ctx, node, eventJobs)
	cp.Watch(ctx)

	<-ctx.Done()

	shutdownctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

}
