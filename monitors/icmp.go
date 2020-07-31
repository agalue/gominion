package monitors

import (
	"fmt"

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
	status := api.PollStatus{}
	if duration, err := tools.Ping(request.IPAddress, request.GetTimeout()); err == nil {
		status.Up(duration.Seconds())
	} else {
		msg := fmt.Sprintf("Error while executing ICMP against %s: %v", request.IPAddress, err)
		status.Down(msg)
	}
	return status
}

var icmpMonitor = &ICMPMonitor{}

func init() {
	RegisterMonitor(icmpMonitor)
}
