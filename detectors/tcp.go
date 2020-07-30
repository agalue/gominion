package detectors

import (
	"fmt"
	"net"

	"github.com/agalue/gominion/api"
)

// TCPDetector represents a detector implementation
type TCPDetector struct {
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (detector *TCPDetector) GetID() string {
	return "TcpDetector"
}

// Detect execute the detector request and return the service status
func (detector *TCPDetector) Detect(request *api.DetectorRequestDTO) api.DetectResults {
	results := api.DetectResults{
		IsServiceDetected: false,
	}

	servAddr := fmt.Sprintf("%s:%s", request.IPAddress, request.GetAttributeValue("port", "23"))
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		return results
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return results
	}
	conn.Close()

	results.IsServiceDetected = true
	return results
}

var tcpDetector = &TCPDetector{}

func init() {
	RegisterDetector(tcpDetector)
}
