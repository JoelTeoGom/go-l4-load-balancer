package event

type Event struct {
	ID     string
	action Action
	Payload
}

type Action string
