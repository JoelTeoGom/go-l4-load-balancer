package apiclient

type RegisterPodRequest struct {
	NodeID      string `json:"nodeId"`
	NodeIP      string `json:"nodeIp"`
	ServiceName string `json:"serviceName"`
}
type ServiceResponse struct {
}
