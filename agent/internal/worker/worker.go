package worker

import (
	"context"
	"fmt"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/event"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/node"
)

func StartWorker(ctx context.Context, node *node.Node, jobQueue <-chan event.Event) error {

	for {
		select {
		case job, ok := <-jobQueue:
			if !ok {
				return nil
			}
			// we dont need to stop the entire agent if some event didnt work
			//TODO give back some feedback to CtrlPlane
			err := processJob(ctx, node, job)
			if err != nil {
				fmt.Printf("Job with ID: %s didnt work as planned [%s, %s]\n", job.ID, job.Action, job.Payload)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func processJob(ctx context.Context, node *node.Node, job event.Event) error {
	switch job.Action {
	case event.ActionCreatePod:
		pod, err := node.CreatePod(job.Payload)
		if err != nil {
			return err
		}
		fmt.Println(pod)
	case event.ActionCreateService:
		service, err := node.CreateService(job.Payload, "IP", "PORT")
		if err != nil {
			return err
		}
		fmt.Println(service)
	}
	return nil
}
