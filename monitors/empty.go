package monitors

// Placeholder for implementing new monitors

import (
	"github.com/agalue/gominion/api"
)

// EmptyMonitor represents a Monitor implementation
type EmptyMonitor struct {
}

// GetID gets the monitor ID (simple class name from its Java counterpart)
func (monitor *EmptyMonitor) GetID() string {
	return "XXX"
}

// Poll execute the XXX monitor request and return the poller response
func (monitor *EmptyMonitor) Poll(request *api.PollerRequestDTO) *api.PollerResponseDTO {
	response := &api.PollerResponseDTO{Status: &api.PollStatus{}, Error: "Not Implemented"}
	response.Status.Unknown("Not implemented")
	return response
}

var emptyMonitor = &EmptyMonitor{}

func init() {
	//	RegisterMonitor(emptyMonitor)
}
