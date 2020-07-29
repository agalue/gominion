package api

import (
	"encoding/xml"
)

// EchoRequest represents an echo request
type EchoRequest struct {
	XMLName     xml.Name `xml:"echo-request"`
	ID          int64    `xml:"id,attr"`
	Message     string   `xml:"message,attr"`
	Body        string   `xml:"body,omitempty"`
	Location    string   `xml:"location,attr"`
	SystemID    string   `xml:"system-id,attr"`
	Delay       int64    `xml:"delay,attr"`
	ShouldThrow bool     `xml:"throw,attr"`
}

// EchoResponse represents an echo response
type EchoResponse struct {
	XMLName xml.Name `xml:"echo-response"`
	ID      int64    `xml:"id,attr"`
	Message string   `xml:"message,attr,omitempty"`
	Body    string   `xml:"body,omitempty"`
	Error   string   `xml:"error,attr,omitempty"`
}
