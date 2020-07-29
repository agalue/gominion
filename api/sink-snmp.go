package api

import "encoding/xml"

// SNMPResults represents a collection of SNMP result instances
type SNMPResults struct {
	Results []SNMPResultDTO `xml:"result"`
}

// TrapIdentityDTO represents the SNMP Trap Identity
type TrapIdentityDTO struct {
	EnterpriseID string `xml:"enterprise-id,attr"`
	Generic      int    `xml:"generic,attr"`
	Specific     int    `xml:"specific,attr"`
}

// TrapDTO represents an SNMP Trap
type TrapDTO struct {
	AgentAddress string           `xml:"agent-address"`
	Community    string           `xml:"community"`
	Version      string           `xml:"version"`
	Timestamp    int64            `xml:"timestamp"`
	CreationTime int64            `xml:"creation-time"`
	PDULength    int              `xml:"pdu-length"`
	RawMessage   []byte           `xml:"raw-message,omitempty"`
	TrapIdentity *TrapIdentityDTO `xml:"trap-identity"`
	Results      *SNMPResults     `xml:"results"`
}

// AddResult adds an SNMP results to the trap object
func (trap *TrapDTO) AddResult(result SNMPResultDTO) {
	if trap.Results == nil {
		trap.Results = &SNMPResults{}
	}
	trap.Results.Results = append(trap.Results.Results, result)
}

// TrapLogDTO represents a collection of SNMP Trap messages
type TrapLogDTO struct {
	XMLName     xml.Name  `xml:"trap-message-log"`
	Location    string    `xml:"location,attr"`
	SystemID    string    `xml:"system-id,attr"`
	TrapAddress string    `xml:"trap-address,attr"`
	Messages    []TrapDTO `xml:"messages"`
}

// AddTrap adds a new trap to the log
func (log *TrapLogDTO) AddTrap(trap TrapDTO) {
	log.Messages = append(log.Messages, trap)
}
