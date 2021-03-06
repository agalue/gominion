package api

import (
	"encoding/xml"
	"strconv"
	"strings"
	"time"
)

// DetectorAttributeDTO represents a detector attribute
type DetectorAttributeDTO struct {
	Key   string `xml:"key,attr"`
	Value string `xml:",chardata"`
}

// DetectorRequestDTO represents a detector request
type DetectorRequestDTO struct {
	XMLName   xml.Name `xml:"detector-request"`
	Location  string   `xml:"location,attr"`
	SystemID  string   `xml:"system-id,attr"`
	ClassName string   `xml:"class-name,attr"`
	IPAddress string   `xml:"address,attr"`

	DetectorAttributes []DetectorAttributeDTO `xml:"detector-attribute,omitempty"`
	RuntimeAttributes  []DetectorAttributeDTO `xml:"runtime-attribute,omitempty"`
}

// GetDetector returns the simple class name for the detector implementation
func (req *DetectorRequestDTO) GetDetector() string {
	if req.ClassName == "" {
		return ""
	}
	sections := strings.Split(req.ClassName, ".")
	return sections[len(sections)-1]
}

// GetTimeout extracts the duration of the timeout attribute if available; otherwise returns default value
func (req *DetectorRequestDTO) GetTimeout() time.Duration {
	if value := req.GetAttributeValueAsInt("timeout"); value > 0 {
		return time.Duration(value) * time.Millisecond
	}
	if value := req.GetRuntimeAttributeValueAsInt("timeout"); value > 0 {
		return time.Duration(value) * time.Millisecond
	}
	return DefaultTimeout
}

// GetRetries extracts the retries attribute if available; otherwise returns default value
func (req *DetectorRequestDTO) GetRetries() int {
	if value := req.GetAttributeValue("retries", ""); value != "" {
		if t, err := strconv.Atoi(value); err != nil {
			return t
		}
	}
	return DefaultRetries
}

// GetAttributeValue extract the value of a given detector attribute
func (req *DetectorRequestDTO) GetAttributeValue(key string, defaultValue string) string {
	if req.DetectorAttributes != nil && len(req.DetectorAttributes) > 0 {
		for _, attr := range req.DetectorAttributes {
			if strings.ToLower(attr.Key) == key {
				return attr.Value
			}
		}
	}
	return defaultValue
}

// GetAttributeValueAsInt extract the value as an integer of a given detector attribute
func (req *DetectorRequestDTO) GetAttributeValueAsInt(key string) int {
	if value := req.GetAttributeValue(key, ""); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return 0
}

// GetRuntimeAttributeValue extract the value of a given runtime attribute
func (req *DetectorRequestDTO) GetRuntimeAttributeValue(key string) string {
	if req.RuntimeAttributes != nil && len(req.RuntimeAttributes) > 0 {
		for _, attr := range req.RuntimeAttributes {
			if strings.ToLower(attr.Key) == key {
				return attr.Value
			}
		}
	}
	return ""
}

// GetRuntimeAttributeValueAsInt extract the value as an integer of a given runtime attribute
func (req *DetectorRequestDTO) GetRuntimeAttributeValueAsInt(key string) int {
	if value := req.GetRuntimeAttributeValue(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return 0
}

// DetectorResponseDTO represents a detector response
type DetectorResponseDTO struct {
	XMLName    xml.Name               `xml:"detector-response"`
	Error      string                 `xml:"error,attr,omitempty"`
	Detected   bool                   `xml:"detected,attr"`
	Attributes []DetectorAttributeDTO `xml:"attribute,omitempty"`
}

// GetStatus returns the detection status as a string
func (response DetectorResponseDTO) GetStatus() string {
	if response.Detected {
		return "detected"
	}
	return "not detected"
}
