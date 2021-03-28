package api

import (
	"encoding/xml"
	"testing"

	"gotest.tools/v3/assert"
)

func TestPollerRequest(t *testing.T) {
	requestXML := `
	<poller-request location="Test" class-name="org.opennms.netmgt.poller.monitors.SnmpMonitor" service-name="SNMP" address="192.168.0.1" node-id="5" node-label="srv01" node-location="Test">
		<attribute key="oid" value=".1.3.6.1.2.1.1.2.0"/>
		<attribute key="retry" value="3"/>
		<attribute key="timeout" value="5000"/>
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

	// Parse and validate Poller request
	request := &PollerRequestDTO{}
	err := xml.Unmarshal([]byte(requestXML), request)
	assert.NilError(t, err)
	assert.Equal(t, "Test", request.Location)
	assert.Equal(t, "SNMP", request.ServiceName)
	assert.Equal(t, "SnmpMonitor", request.GetMonitor())
	assert.Equal(t, ".1.3.6.1.2.1.1.2.0", request.GetAttributeValue("oid", ""))

	// Parse and validate SNMP Agent
	agent := &SNMPAgentDTO{}
	err = xml.Unmarshal([]byte(request.GetAttributeContent("agent")), agent)
	assert.NilError(t, err)
	assert.Equal(t, "192.168.0.1", agent.Address)
	assert.Equal(t, "public", agent.ReadCommunity)
}

func TestPollerResponse(t *testing.T) {
	responseXML := `
	<poller-response>
		<poll-status time="2020-08-05T11:53:01.640732-04:00" response-time="0.0562" code="1" name="Up">
			<properties>
				<property key="response-time">0.0562</property>
			</properties>
		</poll-status>
	</poller-response>
	`
	response := &PollerResponseDTO{}
	err := xml.Unmarshal([]byte(responseXML), response)
	assert.NilError(t, err)
	assert.Equal(t, ServiceAvailableCode, response.Status.StatusCode)
	assert.Equal(t, ServiceAvailable, response.Status.StatusName)
	assert.Equal(t, 2020, response.Status.Timestamp.Year())
	assert.Equal(t, "August", response.Status.Timestamp.Month().String())
	assert.Equal(t, 1, len(response.Status.Properties.PropertyList))
	assert.Equal(t, 0.0562, response.Status.GetPropertyValue("response-time"))
}
