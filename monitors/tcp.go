package monitors

import (
	"fmt"
	"net"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/tools"
)

// TCPMonitor represents a Monitor implementation
type TCPMonitor struct {
}

// GetID gets the monitor ID (simple class name from its Java counterpart)
func (monitor *TCPMonitor) GetID() string {
	return "TcpMonitor"
}

// Poll execute the TCP monitor request and return the the poller response
func (monitor *TCPMonitor) Poll(request *api.PollerRequestDTO) *api.PollerResponseDTO {
	response := &api.PollerResponseDTO{Status: &api.PollStatus{}}
	start := time.Now()
	servAddr := fmt.Sprintf("%s:%s", request.IPAddress, request.GetAttributeValue("port", "23"))
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		response.Status.Down(err.Error())
		return response
	}
	timeout := request.GetTimeout()
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", tcpAddr.String())
	if err != nil {
		response.Status.Down(err.Error())
		return response
	}
	defer conn.Close()

	banner := request.GetAttributeValue("banner", "")
	if ok, err := tools.NetMessageContains(conn, timeout, banner); ok {
		response.Status.Up(time.Since(start).Seconds())
	} else {
		response.Status.Down(err.Error())
	}
	return response
}

func init() {
	RegisterMonitor(&TCPMonitor{})
}
