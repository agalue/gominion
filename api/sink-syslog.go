package api

import (
	"encoding/xml"
)

// SyslogMessageDTO represents a Syslog message
type SyslogMessageDTO struct {
	Timestamp string `xml:"timestamp,attr"`
	Content   []byte `xml:",chardata"`
}

// SyslogMessageLogDTO represents a collection of Syslog messages
type SyslogMessageLogDTO struct {
	XMLName       xml.Name           `xml:"syslog-message-log"`
	SystemID      string             `xml:"system-id,attr"`
	Location      string             `xml:"location,attr"`
	SourceAddress string             `xml:"source-address,attr"`
	SourcePort    int                `xml:"source-port,attr"`
	Messages      []SyslogMessageDTO `xml:"messages"`
}

// AddMessage adds a new message to the log
func (log *SyslogMessageLogDTO) AddMessage(message SyslogMessageDTO) {
	log.Messages = append(log.Messages, message)
}
