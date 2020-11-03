package detectors

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/agalue/gominion/api"
	"github.com/gosnmp/gosnmp"
)

const defaultOID = ".1.3.6.1.2.1.1.2.0"

// SNMPDetector represents a detector implementation
type SNMPDetector struct {
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (detector *SNMPDetector) GetID() string {
	return "SnmpDetector"
}

// Detect execute the SNMP detector request and return the detection response
func (detector *SNMPDetector) Detect(request *api.DetectorRequestDTO) *api.DetectorResponseDTO {
	oid := request.GetAttributeValue("oid", defaultOID)
	isTable := request.GetAttributeValue("isTable", "false")
	matchType := request.GetAttributeValue("matchType", "") // Any, All, None, Exist
	expectedValue := request.GetAttributeValue("vbvalue", "")

	agent := detector.getAgent(request)
	client := agent.GetSNMPClient()
	if err := client.Connect(); err != nil {
		return &api.DetectorResponseDTO{Error: err.Error()}
	}
	defer client.Disconnect()

	return detector.detect(client, oid, matchType, isTable, expectedValue)
}

func (detector *SNMPDetector) detect(client api.SNMPHandler, oid string, matchType string, isTable string, expectedValue string) *api.DetectorResponseDTO {
	response := &api.DetectorResponseDTO{Detected: false}
	if strings.ToLower(isTable) == "true" {
		var returnedValues []string
		err := client.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
			returnedValues = append(returnedValues, fmt.Sprintf("%v", pdu.Value))
			return nil
		})
		if err == nil {
			response = detector.isServiceDetected(matchType, returnedValues, expectedValue)
		} else {
			response.Error = err.Error()
		}
	} else {
		if matchType == "" && expectedValue != "" {
			matchType = "Any"
		}
		if result, err := client.Get(oid); err == nil {
			if len(result.Variables) == 1 {
				value := fmt.Sprintf("%v", result.Variables[0].Value)
				response = detector.isServiceDetected(matchType, []string{value}, expectedValue)
			}
		} else {
			response.Error = err.Error()
		}
	}
	return response
}

func (detector *SNMPDetector) isServiceDetected(matchType string, returnedValues []string, expectedValue string) *api.DetectorResponseDTO {
	response := &api.DetectorResponseDTO{Detected: false}
	if len(returnedValues) == 0 {
		return response
	}
	if matchType == "" {
		matchType = "Exist"
	}
	if matchType != "Exist" && expectedValue == "" {
		response.Error = fmt.Sprintf("expectedValue was not defined using matchType=%s but is required. Otherwise set matchType to Exist", matchType)
		return response
	}
	exp, err := regexp.Compile(expectedValue)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	switch matchType {
	case "Any":
		response.Detected = detector.matchAny(exp, returnedValues)
	case "None":
		response.Detected = !detector.matchAny(exp, returnedValues)
	case "All":
		response.Detected = detector.matchAll(exp, returnedValues)
	case "Exist":
		response.Detected = detector.matchExist(exp, returnedValues)
	}
	return response
}

func (detector *SNMPDetector) matchAny(exp *regexp.Regexp, returnedValues []string) bool {
	for _, returned := range returnedValues {
		if exp.MatchString(returned) {
			return true
		}
	}
	return false
}

func (detector *SNMPDetector) matchAll(exp *regexp.Regexp, returnedValues []string) bool {
	for _, returned := range returnedValues {
		if !exp.MatchString(returned) {
			return false
		}
	}
	return true
}

func (detector *SNMPDetector) matchExist(exp *regexp.Regexp, returnedValues []string) bool {
	for _, returned := range returnedValues {
		if returned != "" {
			return true
		}
	}
	return false
}

func (detector *SNMPDetector) getAgent(request *api.DetectorRequestDTO) *api.SNMPAgentDTO {
	agent := &api.SNMPAgentDTO{
		Address:         request.GetRuntimeAttributeValue("address"),
		Port:            request.GetRuntimeAttributeValueAsInt("port"),
		Timeout:         request.GetRuntimeAttributeValueAsInt("timeout"),
		Retries:         request.GetRuntimeAttributeValueAsInt("retries"),
		MaxVarsPerPdu:   request.GetRuntimeAttributeValueAsInt("max-vars-per-pdu"),
		MaxRepetitions:  request.GetRuntimeAttributeValueAsInt("max-repetitions"),
		MaxRequestSize:  request.GetRuntimeAttributeValueAsInt("max-request-size"),
		Version:         request.GetRuntimeAttributeValueAsInt("version"),
		SecurityLevel:   request.GetRuntimeAttributeValueAsInt("security-level"),
		SecurityName:    request.GetRuntimeAttributeValue("security-name"),
		ReadCommunity:   request.GetRuntimeAttributeValue("read-community"),
		WriteCommunity:  request.GetRuntimeAttributeValue("write-community"),
		AuthPassPhrase:  request.GetRuntimeAttributeValue("auth-passphrase"),
		AuthProtocol:    request.GetRuntimeAttributeValue("auth-protocol"),
		PrivPassPhrase:  request.GetRuntimeAttributeValue("priv-passphrase"),
		PrivProtocol:    request.GetRuntimeAttributeValue("priv-protocol"),
		ContextName:     request.GetRuntimeAttributeValue("context-name"),
		EngineID:        request.GetRuntimeAttributeValue("engine-id"),
		ContextEngineID: request.GetRuntimeAttributeValue("context-engine-id"),
	}
	return agent
}

var snmpDetector = &SNMPDetector{}

func init() {
	RegisterDetector(snmpDetector)
}
