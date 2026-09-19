package reconcile

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/agent"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/event"
)

func StartReconciler(ctx context.Context, node *agent.Agent, jobQueue <-chan event.Event) error {

	for {
		select {
		case job, ok := <-jobQueue:
			if !ok {
				return nil
			}
			// we dont need to stop the entire agent if some event didnt work
			//TODO give back some feedback to CtrlPlane
			err := processJob(node, job)
			if err != nil {
				fmt.Printf("Job with ID: %s didnt work as planned [%s, %s]\n", job.ID, job.Action, job.Payload)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func processJob(node *agent.Agent, job event.Event) error {
	switch job.Action {
	case event.ActionCreatePod:
		servicename := job.Payload
		pod, err := node.CreatePod(servicename)
		if err != nil {
			return err
		}
		fmt.Println(pod)
	case event.ActionCreateService:
		payloadSlice := strings.Split(job.Payload, "+")
		if len(payloadSlice) < 3 {
			return fmt.Errorf("NOT ENOUGH ARGS")
		}
		servicename := payloadSlice[0]
		ip := payloadSlice[1]
		port := payloadSlice[2]
		service, err := node.CreateService(servicename, ip, port)
		if err != nil {
			return err
		}
		fmt.Println(service)
	}
	return nil
}
