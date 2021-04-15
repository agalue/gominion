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

// Poll execute the ICMP monitor request and return the the poller response
func (monitor *ICMPMonitor) Poll(request *api.PollerRequestDTO) *api.PollerResponseDTO {
	response := &api.PollerResponseDTO{Status: &api.PollStatus{}}
	if duration, err := tools.Ping(request.IPAddress, request.GetTimeout()); err == nil {
		response.Status.Up(duration.Seconds())
	} else {
		msg := fmt.Sprintf("Error while executing ICMP against %s: %v", request.IPAddress, err)
		response.Status.Down(msg)
	}
	return response
}

func init() {
	RegisterMonitor(&ICMPMonitor{})
}
