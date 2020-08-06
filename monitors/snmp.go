package monitors

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/soniah/gosnmp"
)

const defaultOID = ".1.3.6.1.2.1.1.2.0"

// SNMPMonitor represents a Monitor implementation
type SNMPMonitor struct {
}

// GetID gets the monitor ID (simple class name from its Java counterpart)
func (monitor *SNMPMonitor) GetID() string {
	return "SnmpMonitor"
}

// Poll execute the monitor request and return the service status
func (monitor *SNMPMonitor) Poll(request *api.PollerRequestDTO) *api.PollerResponseDTO {
	agent := &api.SNMPAgentDTO{}
	response := &api.PollerResponseDTO{Status: &api.PollStatus{}}
	if err := xml.Unmarshal([]byte(request.GetAttributeContent("agent")), agent); err == nil {
		oid := request.GetAttributeValue("oid", defaultOID)
		operator := request.GetAttributeValue("operator", "")
		operand := request.GetAttributeValue("operand", "")
		walkstr := strings.ToLower(request.GetAttributeValue("walk", "false"))
		matchstr := strings.ToLower(request.GetAttributeValue("match-all", "false"))
		minimum := request.GetAttributeValueAsInt("minimum", 0)
		maximum := request.GetAttributeValueAsInt("maximum", 0)
		client := agent.GetSNMPClient()
		if err := client.Connect(); err == nil {
			defer client.Disconnect()
			response = monitor.poll(client, oid, matchstr, walkstr, operator, operand, minimum, maximum)
		} else {
			response.Status.Down(err.Error())
		}
	} else {
		response.Status.Unknown(err.Error())
	}
	return response
}

func (monitor *SNMPMonitor) poll(client api.SNMPHandler, oid string, matchstr string, walkstr string, operator string, operand string, minimum int, maximum int) *api.PollerResponseDTO {
	var response *api.PollerResponseDTO
	if matchstr == "count" {
		response = monitor.processCount(client, oid, operator, operand, minimum, maximum)
	} else if walkstr == "true" {
		response = monitor.processWalk(client, oid, matchstr, operator, operand, minimum, maximum)
	} else {
		response = monitor.processSingle(client, oid, operator, operand, minimum, maximum)
	}
	return response
}

func (monitor *SNMPMonitor) processCount(client api.SNMPHandler, oid string, operator string, operand string, minimum int, maximum int) *api.PollerResponseDTO {
	start := time.Now()
	response := &api.PollerResponseDTO{Status: &api.PollStatus{}}
	if values, err := monitor.walk(client, oid); err == nil {
		count := 0
		for _, value := range values {
			if monitor.meetsCriteria(value, operator, operand) {
				count++
			}
		}
		if count >= minimum && count <= maximum {
			response.Status.Up(time.Since(start).Seconds())
		} else {
			response.Status.Down(fmt.Sprintf("count error: %d not between %d and %d", count, minimum, maximum)) // FIXME
		}
	} else {
		response.Status.Down(err.Error())
	}
	return response
}

func (monitor *SNMPMonitor) processWalk(client api.SNMPHandler, oid string, matchstr string, operator string, operand string, minimum int, maximum int) *api.PollerResponseDTO {
	start := time.Now()
	response := &api.PollerResponseDTO{Status: &api.PollStatus{}}
	if values, err := monitor.walk(client, oid); err == nil {
		for _, value := range values {
			if monitor.meetsCriteria(value, operator, operand) {
				response.Status.Up(time.Since(start).Seconds())
				if matchstr == "false" {
					break
				}
			} else if matchstr == "true" {
				response.Status.Down("something went wrong") // FIXME
				break
			}
		}
	} else {
		response.Status.Down(err.Error())
	}
	return response
}

func (monitor *SNMPMonitor) processSingle(client api.SNMPHandler, oid string, operator string, operand string, minimum int, maximum int) *api.PollerResponseDTO {
	start := time.Now()
	response := &api.PollerResponseDTO{Status: &api.PollStatus{}}
	if result, err := client.Get(oid); err == nil {
		if len(result.Variables) == 1 {
			value := fmt.Sprintf("%v", result.Variables[0].Value)
			if monitor.meetsCriteria(value, operator, operand) {
				response.Status.Up(time.Since(start).Seconds())
			} else {
				response.Status.Down("something went wrong") // FIXME
			}
		} else {
			response.Status.Down("no response")
		}
	} else {
		response.Status.Down(err.Error())
	}
	return response
}

func (monitor *SNMPMonitor) walk(client api.SNMPHandler, oid string) ([]string, error) {
	var returnedValues []string
	err := client.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		returnedValues = append(returnedValues, fmt.Sprintf("%v", pdu.Value))
		return nil
	})
	return returnedValues, err
}

func (monitor *SNMPMonitor) meetsCriteria(result string, operator string, operand string) bool {
	if result != "" && operator != "" && operand != "" {
		if operator == "~" {
			if exp, err := regexp.Compile(operand); err == nil {
				return exp.MatchString(result)
			}
			return false
		}
		intResult, _ := strconv.Atoi(result)
		intOperand, _ := strconv.Atoi(operand)
		switch operator {
		case ">":
			return intResult > intOperand
		case "<":
			return intResult < intOperand
		case ">=":
			return intResult >= intOperand
		case "<=":
			return intResult <= intOperand
		}
		if strings.HasPrefix(operand, ".") {
			result = string([]rune(operand)[1:])
		}
		switch operator {
		case "=":
			return result == operand
		case "!=":
			return result != operand
		}
	}
	return false
}

var snmpMonitor = &SNMPMonitor{}

func init() {
	RegisterMonitor(snmpMonitor)
}
