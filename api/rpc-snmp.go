package api

import (
	"encoding/xml"
	"time"

	"github.com/soniah/gosnmp"
)

// SNMPAgentDTO represents an SNMP agent
// TODO SNMPv3
type SNMPAgentDTO struct {
	Address         string `xml:"address"`
	Version         int    `xml:"version"`
	VersionAsString string `xml:"versionAsString"`
	MaxRepetitions  int    `xml:"maxRepetitions"`
	MaxRequestSize  int    `xml:"maxRequestSize"`
	MaxVarsPerPdu   int    `xml:"maxVarsPerPdu"`
	Port            int    `xml:"port"`
	Retries         int    `xml:"retries"`
	Timeout         int    `xml:"timeout"`
	ReadCommunity   string `xml:"readCommunity"`
	WriteCommunity  string `xml:"writeCommunity"`
	SecurityLevel   int    `xml:"securityLevel"`
	SecurityName    string `xml:"securityName"`
}

// GetSNMPClient gets an SNMP Client instance
// TODO SNMPv3: https://github.com/soniah/gosnmp/blob/master/examples/example3/main.go
func (agent *SNMPAgentDTO) GetSNMPClient() *gosnmp.GoSNMP {
	var version gosnmp.SnmpVersion
	switch agent.Version {
	case 1:
		version = gosnmp.Version1
	case 2:
		version = gosnmp.Version2c
	case 3:
		version = gosnmp.Version3
	}
	return &gosnmp.GoSNMP{
		Target:             agent.Address,
		Port:               uint16(agent.Port),
		Transport:          "udp", // default
		Community:          agent.ReadCommunity,
		Version:            version,
		Timeout:            time.Duration(agent.Timeout) * time.Millisecond,
		ExponentialTimeout: false,
		MaxOids:            60, // default
		Retries:            agent.Retries,
		MaxRepetitions:     uint8(agent.MaxRepetitions),
	}
}

// SNMPGetRequestDTO represents an SNMP get request
type SNMPGetRequestDTO struct {
	XMLName       xml.Name `xml:"get"`
	CorrelationID string   `xml:"correlation-id,attr"`
	OIDs          []string `xml:"oid,omitempty"`
}

// SNMPWalkRequestDTO represents an SNMP walk request
type SNMPWalkRequestDTO struct {
	XMLName        xml.Name `xml:"walk"`
	CorrelationID  string   `xml:"correlation-id,attr"`
	MaxRepetitions int      `xml:"max-repetitions,attr,omitempty"`
	Instance       string   `xml:"instance,attr,omitempty"`
	OIDs           []string `xml:"oid,omitempty"`
}

// SNMPRequestDTO represents an SNMP request
type SNMPRequestDTO struct {
	XMLName     xml.Name             `xml:"snmp-request"`
	Location    string               `xml:"location,attr"`
	SystemID    string               `xml:"system-id,attr"`
	Description string               `xml:"description,attr"`
	Agent       SNMPAgentDTO         `xml:"agent"`
	Gets        []SNMPGetRequestDTO  `xml:"get,omitempty"`
	Walks       []SNMPWalkRequestDTO `xml:"walk,omitempty"`
}

// SNMPValueDTO represents an SNMP value
type SNMPValueDTO struct {
	XMLName xml.Name `xml:"value"`
	Type    int      `xml:"type,attr"`
	Value   string   `xml:",chardata"`
}

// SNMPResultDTO represents an SNMP result
type SNMPResultDTO struct {
	XMLName  xml.Name     `xml:"result"`
	Base     string       `xml:"base"`
	Instance string       `xml:"instance"`
	Value    SNMPValueDTO `xml:"value"`
}

// SNMPResponseDTO represents an SNMP response
type SNMPResponseDTO struct {
	XMLName       xml.Name        `xml:"response"`
	CorrelationID string          `xml:"correlation-id,attr"`
	Results       []SNMPResultDTO `xml:"result"`
}

// SNMPMultiResponseDTO represents an SNMP multi-value response
type SNMPMultiResponseDTO struct {
	XMLName   xml.Name          `xml:"snmp-response"`
	Error     string            `xml:"error,attr,omitempty"`
	Responses []SNMPResponseDTO `xml:"response"`
}
