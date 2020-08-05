package api

import (
	"encoding/xml"
	"time"

	"github.com/soniah/gosnmp"
)

// SNMPHandler represents an SNMP handler based on GoSNMP
type SNMPHandler interface {
	Connect() error
	Disconnect() error
	Version() string
	Target() string
	BulkWalk(rootOid string, walkFn gosnmp.WalkFunc) error
	Get(oid string) (result *gosnmp.SnmpPacket, err error)
}

// SNMPClient represents an SNMP handler implementation
type SNMPClient struct {
	snmp *gosnmp.GoSNMP
}

// Connect initiates a connection against the target device
func (cli *SNMPClient) Connect() error {
	return cli.snmp.Connect()
}

// Disconnect terminates the connection against the target device
func (cli *SNMPClient) Disconnect() error {
	return cli.snmp.Conn.Close()
}

// Version returns the SNMP version
func (cli *SNMPClient) Version() string {
	return cli.snmp.Version.String()
}

// Target returns the target device IP/Hostname
func (cli *SNMPClient) Target() string {
	return cli.snmp.Target
}

// BulkWalk executes an SNMP bulk walk calling WalkFunc after receiving data
func (cli *SNMPClient) BulkWalk(rootOid string, walkFn gosnmp.WalkFunc) error {
	return cli.snmp.BulkWalk(rootOid, walkFn)
}

// Get execute an SNMP GET request
func (cli *SNMPClient) Get(oid string) (result *gosnmp.SnmpPacket, err error) {
	return cli.snmp.Get([]string{oid})
}

// SNMPAgentDTO represents an SNMP agent
type SNMPAgentDTO struct {
	Address         string `xml:"address"`
	ProxyFor        string `xml:"proxyFor"`
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
	AuthPassPhrase  string `xml:"authPassPhrase"`
	AuthProtocol    string `xml:"authProtocol"`
	PrivPassPhrase  string `xml:"privPassPhrase"`
	PrivProtocol    string `xml:"privProtocol"`
	ContextName     string `xml:"contextName"`
	ContextEngineID string `xml:"contextEngineId"`
	EngineID        string `xml:"engineId"`
}

// GetSNMPClient gets an SNMP Client instance
func (agent *SNMPAgentDTO) GetSNMPClient() SNMPHandler {
	session := &gosnmp.GoSNMP{
		Target:             agent.Address,
		Port:               uint16(agent.Port),
		Transport:          "udp", // default
		Community:          agent.ReadCommunity,
		Version:            agent.getVersion(),
		Timeout:            time.Duration(agent.Timeout) * time.Millisecond,
		ExponentialTimeout: false,
		MaxOids:            60, // default
		Retries:            agent.Retries,
		MaxRepetitions:     uint8(agent.MaxRepetitions),
	}
	if agent.Version == 3 {
		session.SecurityModel = gosnmp.UserSecurityModel
		session.MsgFlags = agent.getV3Flags()
		session.SecurityParameters = agent.getSecurityParameters()
	}
	return &SNMPClient{snmp: session}
}

func (agent *SNMPAgentDTO) getVersion() gosnmp.SnmpVersion {
	switch agent.Version {
	case 3:
		return gosnmp.Version3
	case 2:
		return gosnmp.Version2c
	case 1:
		fallthrough
	default:
		return gosnmp.Version1
	}
}

func (agent *SNMPAgentDTO) getV3Flags() gosnmp.SnmpV3MsgFlags {
	switch agent.SecurityLevel {
	case 3:
		return gosnmp.AuthPriv
	case 2:
		return gosnmp.AuthNoPriv
	case 1:
		fallthrough
	default:
		return gosnmp.NoAuthNoPriv
	}
}

func (agent *SNMPAgentDTO) getAuthProtocol() gosnmp.SnmpV3AuthProtocol {
	switch agent.AuthProtocol {
	case "SHA":
		return gosnmp.SHA
	case "MD5":
		fallthrough
	default:
		return gosnmp.MD5
	}
}

func (agent *SNMPAgentDTO) getPrivProtocol() gosnmp.SnmpV3PrivProtocol {
	switch agent.PrivProtocol {
	case "AES":
		return gosnmp.AES
	case "AES192":
		return gosnmp.AES192
	case "AES256":
		return gosnmp.AES256
	case "DES":
		fallthrough
	default:
		return gosnmp.DES
	}
}

func (agent *SNMPAgentDTO) getSecurityParameters() *gosnmp.UsmSecurityParameters {
	params := &gosnmp.UsmSecurityParameters{
		UserName: agent.SecurityName,
	}
	if agent.SecurityLevel > 1 {
		params.AuthenticationPassphrase = agent.AuthPassPhrase
		params.AuthenticationProtocol = agent.getAuthProtocol()
	}
	if agent.SecurityLevel > 2 {
		params.PrivacyPassphrase = agent.PrivPassPhrase
		params.PrivacyProtocol = agent.getPrivProtocol()
	}
	return params
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

// AddResponse adds a response to the multi-response list
func (dto *SNMPMultiResponseDTO) AddResponse(response *SNMPResponseDTO) {
	dto.Responses = append(dto.Responses, *response)
}
