package monitors

import (
	"fmt"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/tools"
)

// ICMPMonitor represents the ICMP Monitor implementation
type ICMPMonitor struct {
}

// GetID gets the monitor ID (simple class name from its Java counterpart)
func (monitor *ICMPMonitor) GetID() string {
	return "IcmpMonitor"
}

// Poll execute the monitor request and return the service status
func (monitor *ICMPMonitor) Poll(request *api.PollerRequestDTO) api.PollStatus {
	var status api.PollStatus
	if duration, err := tools.Ping(request.IPAddress, request.GetTimeout()); err == nil {
		status = api.PollStatus{
			ResponseTime: duration.Seconds(),
			StatusCode:   api.ServiceAvailableCode,
			StatusName:   api.ServiceAvailable,
			Timestamp:    &api.Timestamp{Time: time.Now()},
		}
		status.SetProperty("response-time", status.ResponseTime)
	} else {
		fmt.Printf("Error while executing ICMP against %s: %v", request.IPAddress, err)
		status = api.PollStatus{
			StatusCode: api.ServiceUnavailableCode,
			StatusName: api.ServiceUnavailable,
			Reason:     err.Error(),
		}
	}
	return status
}

var icmpMonitor = &ICMPMonitor{}

func init() {
	RegisterMonitor(icmpMonitor)
}
