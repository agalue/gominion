package rpc

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
	"gotest.tools/assert"
)

func TestSNMP(t *testing.T) {
	requestXML := `
	<snmp-request location="Test" description="SnmpCollectors for 192.168.0.17">
		<agent>
			<maxRepetitions>2</maxRepetitions>
			<maxRequestSize>65535</maxRequestSize>
			<maxVarsPerPdu>10</maxVarsPerPdu>
			<port>161</port>
			<readCommunity>pi4</readCommunity>
			<retries>1</retries>
			<securityLevel>1</securityLevel>
			<securityName>opennmsUser</securityName>
			<timeout>1800</timeout>
			<version>2</version>
			<versionAsString>v2c</versionAsString>
			<writeCommunity>private</writeCommunity>
			<address>192.168.0.17</address>
		</agent>
		<walk correlation-id="0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.2.1</oid>
		</walk>
		<walk correlation-id="1" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.1.3</oid>
		</walk>
		<walk correlation-id="2-0-0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.6.5</oid>
		</walk>
		<walk correlation-id="2-1-0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.6.6</oid>
		</walk>
		<walk correlation-id="2-2-0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.6.7</oid>
		</walk>
		<walk correlation-id="2-3-0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.6.8</oid>
		</walk>
		<walk correlation-id="2-4-0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.6.9</oid>
		</walk>
		<walk correlation-id="2-5-0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.6.10</oid>
		</walk>
		<walk correlation-id="2-6-0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.6.11</oid>
		</walk>
		<walk correlation-id="2-7-0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.6.12</oid>
		</walk>
		<walk correlation-id="2-8-0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.6.14</oid>
		</walk>
		<walk correlation-id="2-9-0" max-repetitions="1" instance=".0">
			<oid>.1.3.6.1.2.1.6.15</oid>
		</walk>
		<walk correlation-id="3-0" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.1</oid>
		</walk>
		<walk correlation-id="3-1" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.15</oid>
		</walk>
		<walk correlation-id="3-2" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.6</oid>
		</walk>
		<walk correlation-id="3-3" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.10</oid>
		</walk>
		<walk correlation-id="3-4" max-repetitions="2">
			<oid>.1.3.6.1.2.1.2.2.1.13</oid>
		</walk>
		<walk correlation-id="3-5" max-repetitions="2">
			<oid>.1.3.6.1.2.1.2.2.1.14</oid>
		</walk>
		<walk correlation-id="3-6" max-repetitions="2">
			<oid>.1.3.6.1.2.1.2.2.1.19</oid>
		</walk>
		<walk correlation-id="3-7" max-repetitions="2">
			<oid>.1.3.6.1.2.1.2.2.1.20</oid>
		</walk>
		<walk correlation-id="3-8" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.7</oid>
		</walk>
		<walk correlation-id="3-9" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.8</oid>
		</walk>
		<walk correlation-id="3-10" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.9</oid>
		</walk>
		<walk correlation-id="3-11" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.11</oid>
		</walk>
		<walk correlation-id="3-12" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.12</oid>
		</walk>
		<walk correlation-id="3-13" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.13</oid>
		</walk>
		<walk correlation-id="3-14" max-repetitions="2">
			<oid>.1.3.6.1.2.1.31.1.1.1.18</oid>
		</walk>
	</snmp-request>
	`
	request := &ipc.RpcRequestProto{
		Location:   "Test",
		ModuleId:   "SNMP",
		SystemId:   "minion01",
		RpcId:      "0001",
		RpcContent: []byte(requestXML),
	}
	response := snmpModule.Execute(request)
	snmpResponse := &api.SNMPMultiResponseDTO{}
	err := xml.Unmarshal(response.RpcContent, snmpResponse)
	assert.NilError(t, err)
	bytes, err := xml.MarshalIndent(snmpResponse, "", "  ")
	fmt.Println(string(bytes))
}
