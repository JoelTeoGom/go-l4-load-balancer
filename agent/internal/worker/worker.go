package worker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/event"
	"github.com/JoelTeoGom/go-l4-load-balancer/agent/internal/node"
)

type Worker struct {
	Node         *node.Node
	StreamClient Client
}

type Client struct {
	url         string
	http        *http.Client
	idleTimeout time.Duration // has to be  >  keepalive interval coming from server
	retry       time.Duration
	lastID      string
}

func NewWorker(node *node.Node, url string) *Worker {
	return &Worker{
		Node: node,
		StreamClient: Client{
			http: &http.Client{ //!!!!!NO GLOBAL TIMEOUT: would kill stream
				Transport: &http.Transport{
					DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
					TLSHandshakeTimeout:   5 * time.Second,
					ResponseHeaderTimeout: 10 * time.Second, // only until we receive headers
				},
			},
			url:         url,
			idleTimeout: 45 * time.Second, // 3x  keepalive COMPARED FROM  15s server
			retry:       3 * time.Second,  //SSE SPEC DEFAULTS
		},
	}
}

func (w *Worker) StartWorker(ctx context.Context, jobQueue <-chan event.Event) error {

	for {
		select {
		case job, ok := <-jobQueue:
			if !ok {
				return
			}
			// we dont need to stop the entire agent if some event didnt work
			//TODO give back some feedback to CtrlPlane
			err := processJob(ctx, node, job)
			if err != nil {
				fmt.Printf("Job with ID: %s didnt work as planned [%s, %s]\n", job.ID, job.Action, job.Payload)
			}
		case <-ctx.Done():
			return
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
