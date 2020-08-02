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

// ServiceUnknown poll status name unknown
const ServiceUnknown = "Unknown"

// ServiceUnknownCode poll status cod for ServiceUnknown
const ServiceUnknownCode = 0

// ServiceAvailable poll status name available
const ServiceAvailable = "Up"

// ServiceAvailableCode poll status cod for ServiceAvailable
const ServiceAvailableCode = 1

// ServiceUnavailable poll status name unavailable
const ServiceUnavailable = "Down"

// ServiceUnavailableCode poll status cod for ServiceUnavailable
const ServiceUnavailableCode = 2

// ServiceUnresponsive poll status name unavailable
const ServiceUnresponsive = "Unresponsive"

// ServiceUnresponsiveCode poll status cod for ServiceUnresponsive
const ServiceUnresponsiveCode = 3

// PollStatusProperties represents a poll status property list
type PollStatusProperties struct {
	XMLName    xml.Name `xml:"properties"`
	Properties []PollStatusProperty
}

// PollStatusProperty represents a poll status property
type PollStatusProperty struct {
	XMLName xml.Name `xml:"property"`
	Key     string   `xml:"key,attr"`
	Value   float64  `xml:",chardata"`
}

// PollStatus represents a poll status
type PollStatus struct {
	XMLName      xml.Name              `xml:"poll-status"`
	Timestamp    *Timestamp            `xml:"time,attr,omitempty"`
	Reason       string                `xml:"reason,attr,omitempty"`
	ResponseTime float64               `xml:"response-time,attr"`
	StatusCode   int                   `xml:"code,attr"`
	StatusName   string                `xml:"name,attr"`
	Properties   *PollStatusProperties `xml:"properties,omitempty"`
}

// Up update the poll status for an available service
func (status *PollStatus) Up(responseTime float64) *PollStatus {
	status.StatusCode = ServiceAvailableCode
	status.StatusName = ServiceAvailable
	status.Timestamp = &Timestamp{Time: time.Now()}
	status.ResponseTime = responseTime
	status.SetProperty("response-time", responseTime)
	return status
}

// Down update the poll status for an unavailable service
func (status *PollStatus) Down(reason string) {
	status.StatusCode = ServiceUnavailableCode
	status.StatusName = ServiceUnavailable
	status.Timestamp = &Timestamp{Time: time.Now()}
	status.Reason = reason
}

// Unknown update the poll status to be unknown
func (status *PollStatus) Unknown(reason string) {
	status.StatusCode = ServiceUnknownCode
	status.StatusName = ServiceUnknown
	status.Timestamp = &Timestamp{Time: time.Now()}
	status.Reason = reason
}

// SetProperty adds or updates an existing property
func (status *PollStatus) SetProperty(key string, value float64) {
	if status.Properties == nil {
		status.Properties = &PollStatusProperties{}
	}
	found := false
	for _, p := range status.Properties.Properties {
		if p.Key == key {
			p.Value = value
			found = true
		}
	}
	if !found {
		p := PollStatusProperty{Key: key, Value: value}
		status.Properties.Properties = append(status.Properties.Properties, p)
	}
}

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
	XMLName xml.Name    `xml:"poller-response"`
	Error   string      `xml:"error,attr,omitempty"`
	Status  *PollStatus `xml:"poll-status"`
}
