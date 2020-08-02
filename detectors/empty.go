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
	return "EMPTY"
}

// Detect execute the detector request and return the service status
func (detector *EmptyDetector) Detect(request *api.DetectorRequestDTO) *api.DetectorResponseDTO {
	response := &api.DetectorResponseDTO{Detected: false}
	return response
}

var emptyDetector = &EmptyDetector{}

func init() {
	//	RegisterDetector(emptyDetector)
}
