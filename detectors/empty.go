package detectors

// Placeholder for implementing new detectors

import (
	"github.com/agalue/gominion/api"
)

// EmptyDetector represents a detector implementation
type EmptyDetector struct {
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (detector *EmptyDetector) GetID() string {
	return "XXX"
}

// Detect execute the XXX detector request and return the detection response
func (detector *EmptyDetector) Detect(request *api.DetectorRequestDTO) *api.DetectorResponseDTO {
	response := &api.DetectorResponseDTO{Detected: false, Error: "Not Implemented"}
	return response
}

var emptyDetector = &EmptyDetector{}

func init() {
	//	RegisterDetector(emptyDetector)
}
