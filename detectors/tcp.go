package detectors

import (
	"fmt"
	"net"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/tools"
)

// TCPDetector represents a detector implementation
type TCPDetector struct {
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (detector *TCPDetector) GetID() string {
	return "TcpDetector"
}

// Detect execute the detector request and return the service status
func (detector *TCPDetector) Detect(request *api.DetectorRequestDTO) *api.DetectorResponseDTO {
	results := &api.DetectorResponseDTO{Detected: false}

	servAddr := fmt.Sprintf("%s:%s", request.IPAddress, request.GetAttributeValue("port", "23"))
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		results.Error = err.Error()
		return results
	}

	timeout := request.GetTimeout()
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", tcpAddr.String())
	if err != nil {
		results.Error = err.Error()
		return results
	}
	defer conn.Close()

	banner := request.GetAttributeValue("banner", "")
	if ok, err := tools.NetMessageContains(conn, timeout, banner); ok {
		results.Detected = true
	} else {
		results.Error = err.Error()
	}
	return results
}

var tcpDetector = &TCPDetector{}

func init() {
	RegisterDetector(tcpDetector)
}
