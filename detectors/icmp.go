package detectors

import (
	"log"

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

// Detect execute the detector request and return the service status
func (detector *ICMPDetector) Detect(request *api.DetectorRequestDTO) api.DetectResults {
	results := api.DetectResults{}
	if _, err := tools.Ping(request.IPAddress, request.GetTimeout()); err == nil {
		results.IsServiceDetected = true
	} else {
		log.Printf("Error during detection: %v", err)
		results.IsServiceDetected = false
	}
	return results
}

var icmpDetector = &ICMPDetector{}

func init() {
	RegisterDetector(icmpDetector)
}
