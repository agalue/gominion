package api

import (
	"encoding/xml"
	"strconv"
	"strings"
	"time"
)

// DefaultTimeout default timeout duration
const DefaultTimeout = 3 * time.Second

// DefaultRetries default retries
const DefaultRetries = 2

// PollerAttributeDTO represents a poller atrribute
type PollerAttributeDTO struct {
	Key     string `xml:"key,attr"`
	Value   string `xml:"value,attr"`
	Content string `xml:",innerxml"`
}

// PollerRequestDTO represents a poller request
type PollerRequestDTO struct {
	XMLName      xml.Name             `xml:"poller-request"`
	Location     string               `xml:"location,attr"`
	SystemID     string               `xml:"system-id,attr"`
	ClassName    string               `xml:"class-name,attr"`
	ServiceName  string               `xml:"service-name,attr"`
	IPAddress    string               `xml:"address,attr"`
	NodeID       string               `xml:"node-id,attr"`
	NodeLabel    string               `xml:"node-label,attr"`
	NodeLocation string               `xml:"node-location,attr"`
	Attributes   []PollerAttributeDTO `xml:"attribute,omitempty"`
}

// GetTimeout extracts the duration of the timeout attribute if available; otherwise returns default value
func (req *PollerRequestDTO) GetTimeout() time.Duration {
	if value := req.GetAttributeValue("timeout", ""); value != "" {
		if t, err := strconv.Atoi(value); err != nil {
			return time.Duration(t) * time.Microsecond
		}
	}
	return DefaultTimeout
}

// GetRetries extracts the retries attribute if available; otherwise returns default value
func (req *PollerRequestDTO) GetRetries() int {
	if value := req.GetAttributeValue("retries", ""); value != "" {
		if t, err := strconv.Atoi(value); err != nil {
			return t
		}
	}
	return DefaultRetries
}

// GetMonitor returns the simple class name for the monitor implementation
func (req *PollerRequestDTO) GetMonitor() string {
	if req.ClassName == "" {
		return ""
	}
	sections := strings.Split(req.ClassName, ".")
	return sections[len(sections)-1]
}

// GetAttributeValue gets the value of a given attribute
func (req *PollerRequestDTO) GetAttributeValue(key string, defaultValue string) string {
	if req.Attributes != nil && len(req.Attributes) > 0 {
		for _, attr := range req.Attributes {
			if strings.ToLower(attr.Key) == key {
				return attr.Value
			}
		}
	}
	return defaultValue
}

// GetAttributeContent gets the value of a given attribute
func (req *PollerRequestDTO) GetAttributeContent(key string) string {
	if req.Attributes != nil && len(req.Attributes) > 0 {
		for _, attr := range req.Attributes {
			if strings.ToLower(attr.Key) == key {
				return attr.Content
			}
		}
	}
	return ""
}

// PollerResponseDTO represents a poller response
type PollerResponseDTO struct {
	XMLName xml.Name   `xml:"poller-response"`
	Error   string     `xml:"error,attr,omitempty"`
	Status  PollStatus `xml:"poll-status"`
}
