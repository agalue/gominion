package detectors

import (
	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/tools"
)

// ICMPDetector represents a detector implementation
type ICMPDetector struct {
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (detector *ICMPDetector) GetID() string {
	return "IcmpDetector"
}

// Detect execute the ICMP detector request and return the detection response
func (detector *ICMPDetector) Detect(request *api.DetectorRequestDTO) *api.DetectorResponseDTO {
	results := &api.DetectorResponseDTO{}
	if _, err := tools.Ping(request.IPAddress, request.GetTimeout()); err == nil {
		results.Detected = true
	} else {
		results.Error = err.Error()
		results.Detected = false
	}
	return results
}

var icmpDetector = &ICMPDetector{}

func init() {
	RegisterDetector(icmpDetector)
}
