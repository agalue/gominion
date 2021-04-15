package monitors

import (
	"github.com/agalue/gominion/api"
)

// JMXMonitor represents the JMX Monitor implementation
type JMXMonitor struct {
}

// GetID gets the monitor ID (simple class name from its Java counterpart)
func (monitor *JMXMonitor) GetID() string {
	return "Jsr160Monitor"
}

// Poll execute the JMX monitor request and return the the poller response.
// It only assumes the service is up for JMX-Minion; returns an error otherwise.
func (monitor *JMXMonitor) Poll(request *api.PollerRequestDTO) *api.PollerResponseDTO {
	response := &api.PollerResponseDTO{Status: &api.PollStatus{}}
	// Whitelist JMX-Minion by default to avoid outages.
	if request.ServiceName == "JMX-Minion" && request.IPAddress == "127.0.0.1" {
		response.Status.Up(0.0)
	} else {
		response.Status.Unknown("JMX not supported by gominion")
	}
	return response
}

func init() {
	RegisterMonitor(&JMXMonitor{})
}
