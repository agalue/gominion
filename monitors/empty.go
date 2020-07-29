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
	return "EMPTY"
}

// Poll execute the monitor request and return the service status
func (monitor *EmptyMonitor) Poll(request *api.PollerRequestDTO) api.PollStatus {
	status := api.PollStatus{
		StatusCode: api.ServiceUnknownCode,
		StatusName: api.ServiceUnknown,
	}
	return status
}

var emptyMonitor = &EmptyMonitor{}

func init() {
	//	RegisterMonitor(emptyMonitor)
}
