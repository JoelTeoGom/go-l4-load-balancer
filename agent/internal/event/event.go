package event

type Event struct {
	ID      string
	Action  Action
	Payload string
}

type Action string

const (
	ActionCreatePod     Action = "CREATE_POD"
	ActionCreateService Action = "CREATE_SERVICE"
)
