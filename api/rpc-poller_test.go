package api

import (
	"encoding/xml"
	"fmt"
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestPollerRequest(t *testing.T) {
	request := &PollerRequestDTO{
		Location:     "Test",
		SystemID:     "minion1",
		ClassName:    "org.opennms.netmgt.poller.monitors.IcmpMonitor",
		ServiceName:  "ICMP",
		IPAddress:    "127.0.0.1",
		NodeID:       "1",
		NodeLabel:    "srv01",
		NodeLocation: "Test",
		Attributes: []PollerAttributeDTO{
			{
				Key:   "retry",
				Value: "3",
			},
			{
				Key:   "timeout",
				Value: "5000",
			},
		},
	}
	bytes, err := xml.MarshalIndent(request, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
	assert.Equal(t, "IcmpMonitor", request.GetMonitor())
}

func TestPollerRequestUnmarshal(t *testing.T) {
	requestXML := `
	<poller-request location="Test" class-name="org.opennms.netmgt.poller.monitors.SnmpMonitor" service-name="SNMP" address="192.168.0.1" node-id="5" node-label="srv01" node-location="Test">
		<attribute key="oid" value=".1.3.6.1.2.1.1.2.0"/>
		<attribute key="agent">
		<snmpAgentConfig>
			<maxRepetitions>2</maxRepetitions>
			<maxRequestSize>65535</maxRequestSize>
			<maxVarsPerPdu>10</maxVarsPerPdu>
			<port>161</port>
			<readCommunity>public</readCommunity>
			<retries>1</retries>
			<securityLevel>1</securityLevel>
			<securityName>opennmsUser</securityName>
			<timeout>1800</timeout>
			<version>2</version>
			<versionAsString>v2c</versionAsString>
			<writeCommunity>private</writeCommunity>
			<address>192.168.0.1</address>
		</snmpAgentConfig>
		</attribute>
	</poller-request>
	`
	request := &PollerRequestDTO{}
	err := xml.Unmarshal([]byte(requestXML), request)
	assert.NilError(t, err)
	assert.Equal(t, ".1.3.6.1.2.1.1.2.0", request.GetAttributeValue("oid", ""))
	agent := &SNMPAgentDTO{}
	err = xml.Unmarshal([]byte(request.GetAttributeContent("agent")), agent)
	assert.NilError(t, err)
	assert.Equal(t, "192.168.0.1", agent.Address)
}

func TestPollerResponse(t *testing.T) {
	response := &PollerResponseDTO{
		Status: &PollStatus{
			StatusCode:   ServiceAvailableCode,
			StatusName:   ServiceAvailable,
			ResponseTime: 0.0562,
			Timestamp:    &Timestamp{Time: time.Now()},
			Properties: &PollStatusProperties{
				Properties: []PollStatusProperty{
					{
						Key:   "response-time",
						Value: 0.0561999999,
					},
				},
			},
		},
	}
	bytes, err := xml.MarshalIndent(response, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
}

func TestPollerResponseEmpty(t *testing.T) {
	response := &PollerResponseDTO{
		Status: &PollStatus{
			StatusCode: ServiceUnknownCode,
			StatusName: ServiceUnknown,
		},
	}
	bytes, err := xml.MarshalIndent(response, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
}
