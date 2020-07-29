package monitors

import (
	"encoding/xml"
	"log"
	"time"

	"github.com/agalue/gominion/api"
)

// SNMPMonitor represents a Monitor implementation
type SNMPMonitor struct {
}

// GetID gets the monitor ID (simple class name from its Java counterpart)
func (monitor *SNMPMonitor) GetID() string {
	return "SnmpMonitor"
}

// Poll execute the monitor request and return the service status
// TODO Implement all pending features
func (monitor *SNMPMonitor) Poll(request *api.PollerRequestDTO) api.PollStatus {
	oid := request.GetAttributeValue("oid")
	agent := &api.SNMPAgentDTO{}

	status := api.PollStatus{
		StatusCode: api.ServiceUnavailableCode,
		StatusName: api.ServiceUnavailable,
	}

	if err := xml.Unmarshal([]byte(request.GetAttributeContent("agent")), agent); err == nil {
		log.Printf("Requesting OID %s from %s", oid, agent.Address)
		start := time.Now()
		client := agent.GetSNMPClient()
		if err := client.Connect(); err == nil {
			defer client.Conn.Close()
			if _, err := client.Get([]string{oid}); err == nil {
				duration := time.Since(start)
				status.StatusCode = api.ServiceAvailableCode
				status.StatusName = api.ServiceAvailable
				status.ResponseTime = duration.Seconds()
				status.SetProperty("response-time", status.ResponseTime)
			} else {
				status.Reason = err.Error()
			}
		} else {
			status.Reason = err.Error()
		}
	} else {
		status.Reason = err.Error()
	}

	return status
}

var snmpMonitor = &SNMPMonitor{}

func init() {
	RegisterMonitor(snmpMonitor)
}
