package controlplane

type Event struct {
	ID      string
	action  Action
	Payload string
}

type Action string

const (
	ActionCreatePod     Action = "CREATE_POD"
	ActionCreateService Action = "CREATE_SERVICE"
)
