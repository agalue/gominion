package api

// DetectResults represents a detector results
type DetectResults struct {
	IsServiceDetected bool
	ServiceAttributes map[string]string
}

// ServiceDetector represents the service detector interface
type ServiceDetector interface {
	GetID() string
	Detect(request *DetectorRequestDTO) DetectResults
}
