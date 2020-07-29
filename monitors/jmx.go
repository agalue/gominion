package monitors

import (
	"time"

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
	var status api.PollStatus
	if request.ServiceName == "JMX-Minion" && request.IPAddress == "127.0.0.1" {
		status = api.PollStatus{
			ResponseTime: 1.0,
			StatusCode:   api.ServiceAvailableCode,
			StatusName:   api.ServiceAvailable,
			Timestamp:    &api.Timestamp{Time: time.Now()},
		}
		status.SetProperty("response-time", status.ResponseTime)
	} else {
		status = api.PollStatus{
			StatusCode: api.ServiceUnknownCode,
			StatusName: api.ServiceUnknown,
			Reason:     "JMX not supported by gominion",
		}
	}
	return status
}

var jmxMonitor = &JMXMonitor{}

func init() {
	RegisterMonitor(jmxMonitor)
}
