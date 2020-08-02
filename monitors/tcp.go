package monitors

import (
	"fmt"
	"net"
	"time"

	"github.com/agalue/gominion/api"
)

// TCPMonitor represents a Monitor implementation
type TCPMonitor struct {
}

// GetID gets the monitor ID (simple class name from its Java counterpart)
func (monitor *TCPMonitor) GetID() string {
	return "TcpMonitor"
}

// Poll execute the monitor request and return the service status
func (monitor *TCPMonitor) Poll(request *api.PollerRequestDTO) *api.PollerResponseDTO {
	response := &api.PollerResponseDTO{Status: &api.PollStatus{}}
	start := time.Now()
	servAddr := fmt.Sprintf("%s:%s", request.IPAddress, request.GetAttributeValue("port", "23"))
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		response.Status.Down(err.Error())
		return response
	}
	dialer := net.Dialer{Timeout: request.GetTimeout()}
	conn, err := dialer.Dial("tcp", tcpAddr.String())
	if err != nil {
		response.Status.Down(err.Error())
		return response
	}
	conn.Close()
	duration := time.Since(start)
	response.Status.Up(duration.Seconds())
	return response
}

var tcpMonitor = &TCPMonitor{}

func init() {
	RegisterMonitor(tcpMonitor)
}
