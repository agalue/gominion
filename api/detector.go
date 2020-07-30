package api

// DetectResults represents a detector results
type DetectResults struct {
	IsServiceDetected bool
	ServiceAttributes map[string]string
}

// AddAttribute adds attribute to the detector results
func (result *DetectResults) AddAttribute(key string, value string) {
	if result.ServiceAttributes == nil {
		result.ServiceAttributes = make(map[string]string)
	}
	result.ServiceAttributes[key] = value
}

// ServiceDetector represents the service detector interface
type ServiceDetector interface {
	GetID() string
	Detect(request *DetectorRequestDTO) DetectResults
}
