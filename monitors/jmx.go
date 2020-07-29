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

// Poll execute the monitor request and return the service status
func (monitor *JMXMonitor) Poll(request *api.PollerRequestDTO) api.PollStatus {
	status := api.PollStatus{}
	// Whitelist JMX-Minion by default to avoid outages.
	if request.ServiceName == "JMX-Minion" && request.IPAddress == "127.0.0.1" {
		status.Up(0.0)
	} else {
		status.Unknown("JMX not supported by gominion")
	}
	return status
}

var jmxMonitor = &JMXMonitor{}

func init() {
	RegisterMonitor(jmxMonitor)
}
