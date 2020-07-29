package detectors

import (
	"log"

	"github.com/agalue/gominion/api"
)

const defaultOID = ".1.3.6.1.2.1.1.2.0"

// SNMPDetector represents a detector implementation
type SNMPDetector struct {
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (detector *SNMPDetector) GetID() string {
	return "SnmpDetector"
}

// Detect execute the detector request and return the service status
func (detector *SNMPDetector) Detect(request *api.DetectorRequestDTO) api.DetectResults {
	results := api.DetectResults{
		IsServiceDetected: false,
	}

	agent := detector.getAgent(request)
	client := agent.GetSNMPClient()
	if err := client.Connect(); err != nil {
		log.Printf("Connect Error: %v\n", err)
		return results
	}
	defer client.Conn.Close()

	oid := request.GetAttributeValue("oid", defaultOID)
	if _, err := client.Get([]string{oid}); err == nil {
		results.IsServiceDetected = true
	}

	return results
}

/*
  TODO
  <runtime-attribute key="auth-passphrase"></runtime-attribute>
  <runtime-attribute key="auth-protocol"></runtime-attribute>
  <runtime-attribute key="priv-passphrase"></runtime-attribute>
  <runtime-attribute key="priv-protocol"></runtime-attribute>
  <runtime-attribute key="context-name"></runtime-attribute>
  <runtime-attribute key="engine-id"></runtime-attribute>
  <runtime-attribute key="context-engine-id"></runtime-attribute>
  <runtime-attribute key="enterprise-id"></runtime-attribute>
*/
func (detector *SNMPDetector) getAgent(request *api.DetectorRequestDTO) *api.SNMPAgentDTO {
	agent := &api.SNMPAgentDTO{
		Address:        request.GetRuntimeAttributeValue("address"),
		Port:           request.GetRuntimeAttributeValueAsInt("port"),
		Timeout:        request.GetRuntimeAttributeValueAsInt("timeout"),
		Retries:        request.GetRuntimeAttributeValueAsInt("retries"),
		MaxVarsPerPdu:  request.GetRuntimeAttributeValueAsInt("max-vars-per-pdu"),
		MaxRepetitions: request.GetRuntimeAttributeValueAsInt("max-repetitions"),
		MaxRequestSize: request.GetRuntimeAttributeValueAsInt("max-request-size"),
		Version:        request.GetRuntimeAttributeValueAsInt("version"),
		SecurityLevel:  request.GetRuntimeAttributeValueAsInt("security-level"),
		SecurityName:   request.GetRuntimeAttributeValue("security-name"),
		ReadCommunity:  request.GetRuntimeAttributeValue("read-community"),
		WriteCommunity: request.GetRuntimeAttributeValue("write-community"),
	}
	return agent
}

var snmpDetector = &SNMPDetector{}

func init() {
	RegisterDetector(snmpDetector)
}
