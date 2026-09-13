package registry

type Status string

const (
	// StatusActive indicates that the node is active and healthy.
	StatusActive Status = "active"

	// StatusInactive indicates that the node is inactive or unhealthy.
	StatusInactive Status = "inactive"
)
