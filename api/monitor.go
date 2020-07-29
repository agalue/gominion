package api

import (
	"encoding/xml"
	"time"
)

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

// ServiceMonitor represents the service monitor interface
type ServiceMonitor interface {
	GetID() string
	Poll(request *PollerRequestDTO) PollStatus
}
