package api

import (
	"encoding/xml"
)

// DNSLookupRequestDTO represents a DNS Lookup request
type DNSLookupRequestDTO struct {
	XMLName     xml.Name `xml:"dns-lookup-request"`
	Location    string   `xml:"location,attr"`
	SystemID    string   `xml:"system-id,attr"`
	HostRequest string   `xml:"host-request,attr"`
	QueryType   string   `xml:"query-type,attr"`
}

// DNSLookupResponseDTO represents a DNS Lookup response
type DNSLookupResponseDTO struct {
	XMLName      xml.Name `xml:"dns-lookup-response"`
	HostResponse string   `xml:"host-response,attr,omitempty"`
	Error        string   `xml:"error,attr,omitempty"`
}
