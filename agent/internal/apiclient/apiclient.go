package apiclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/JoelTeoGom/kubernetes-from-scratch/agent/internal/agent"
	"github.com/JoelTeoGom/kubernetes-from-scratch/agent/internal/event"
)

// APIClient is the HTTP client the agent uses to talk to the orchestrator's API server.
type APIClient struct {
	Client       *http.Client
	StreamClient StreamClient
	Node         *agent.Agent
	EventQueue   chan<- event.Event
}

// StreamClient is the SSE-tuned client used to keep the watch stream open.
type StreamClient struct {
	url         string
	http        *http.Client
	idleTimeout time.Duration // has to be  >  keepalive interval coming from server
	retry       time.Duration
	lastID      string
}

func NewAPIClient(node *agent.Agent, eventQueue chan<- event.Event) *APIClient {
	return &APIClient{
		EventQueue: eventQueue,
		Node:       node,
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		StreamClient: StreamClient{
			http: &http.Client{ //!!!!!NO GLOBAL TIMEOUT: would kill stream
				Transport: &http.Transport{
					DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
					TLSHandshakeTimeout:   5 * time.Second,
					ResponseHeaderTimeout: 10 * time.Second, // only until we receive headers
				},
			},
			url:         node.ControlPlaneURL,
			idleTimeout: 45 * time.Second, // 3x  keepalive COMPARED FROM  15s server
			retry:       3 * time.Second,  //SSE SPEC DEFAULTS
		},
	}
}

func (cp *APIClient) RegisterNode(ctx context.Context) error {
	return cp.post(ctx, "/register-node", cp.Node.NodeID)
}

func (cp *APIClient) List(ctx context.Context) error {
	endpoint := cp.Node.ControlPlaneURL + "/list/" + cp.Node.NodeID
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := cp.Client.Do(request)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, response.Body) // drain so the conn can be reused
		response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		errorBody, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return fmt.Errorf("unexpected status %d: %s", response.StatusCode, errorBody)
	}

	var serviceResponse ServiceResponse
	if err := json.NewDecoder(response.Body).Decode(&serviceResponse); err != nil {
		return nil, fmt.Errorf("Error decoding serviceResponse %w", err)
	}
	return &serviceResponse, nil
}

func (cp *APIClient) UnregisterNode(ctx context.Context) error {
	return cp.post(ctx, "/unregister-node", cp.Node.NodeID)
}

func (cp *APIClient) RegisterPod(ctx context.Context, serviceName string) error {
	return cp.post(ctx, "/register-pod", RegisterPodRequest{
		NodeID:      cp.Node.NodeID,
		NodeIP:      cp.Node.NodeIP,
		ServiceName: serviceName,
	})
}

func (cp *APIClient) post(ctx context.Context, path string, payload any) error {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	endpoint := cp.Node.ControlPlaneURL + path
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := cp.Client.Do(request)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, response.Body) // drain so the conn can be reused
		response.Body.Close()
	}()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		errorBody, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return fmt.Errorf("unexpected status %d: %s", response.StatusCode, errorBody)
	}
	return nil
}

func (cp *APIClient) Watch(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cp.StreamClient.url+"/watch-node", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if cp.StreamClient.lastID != "" {
		req.Header.Set("Last-Event-ID", cp.StreamClient.lastID) // TODO server can start FROM LAST ID STORED
	}
	resp, err := cp.StreamClient.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error connecting watchdog to ctrlPlane %s", cp.Node.NodeID)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		return fmt.Errorf("unexpected content-type %q", ct)
	}

	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		//Doing this For Fun :)
		//ex: "id,action,payload"
		strline := string(line)
		eventline := strings.Split(strline, ",")
		if len(eventline) < 3 {
			continue //NOW WE ONLY ACCEPT EVENTS
		}
		event := event.Event{
			ID:      eventline[0],
			Action:  event.Action(eventline[1]),
			Payload: eventline[2],
		}

		sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		//IF WE CANNOT SCHEDULE THE EVENT IN 5 SEC WE DISCARD IT (in the future we could do like kubernetes and have
		// the events stored in db and reconcile the state using a worker with a ticker)
		select {
		case cp.EventQueue <- event:
		case <-sendCtx.Done():
		}
		cancel()
	}
	return nil
}
