package api

import (
	"encoding/xml"
	"time"
)

// PingRequest represents a ping request
type PingRequest struct {
	XMLName    xml.Name `xml:"ping-request"`
	Location   string   `xml:"location,attr"`
	SystemID   string   `xml:"system-id,attr"`
	Retries    int      `xml:"retries,attr"`
	Timeout    int      `xml:"timeout,attr"`
	Address    string   `xml:"address"`
	PacketSize int      `xml:"packet-size"`
}

// GetTimeout gets the timeout duration
func (req *PingRequest) GetTimeout() time.Duration {
	return time.Duration(req.Timeout) * time.Microsecond
}

// PingResponse represents a ping response
type PingResponse struct {
	XMLName xml.Name `xml:"ping-response"`
	RTT     float64  `xml:"rtt,omitempty"`
	Error   string   `xml:"error,attr,omitempty"`
}
