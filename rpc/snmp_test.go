package rpc

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/tools"
	"github.com/gosnmp/gosnmp"
	"gotest.tools/v3/assert"
)

var requestXML = `<snmp-request location="Test" description="SnmpCollectors for 192.168.0.17">
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
		<oid>.1.3.6.1.2.1.2.1</oid> <!-- IF-MIB::ifNumber -->
	</walk>
	<walk correlation-id="1" max-repetitions="1" instance=".0">
		<oid>.1.3.6.1.2.1.1.3</oid> <!-- SNMPv2-MIB::sysUpTime -->
	</walk>
	<walk correlation-id="2-0-0" max-repetitions="1" instance=".0">
		<oid>.1.3.6.1.2.1.6.5</oid> <!-- TCP-MIB::tcpActiveOpens -->
	</walk>
	<walk correlation-id="2-1-0" max-repetitions="1" instance=".0">
		<oid>.1.3.6.1.2.1.6.6</oid> <!-- TCP-MIB::tcpPassiveOpens -->
	</walk>
	<walk correlation-id="3-0" max-repetitions="2">
		<oid>.1.3.6.1.2.1.31.1.1.1.1</oid> <!-- IF-MIB::ifName -->
	</walk>
	<walk correlation-id="3-1" max-repetitions="2">
		<oid>.1.3.6.1.2.1.31.1.1.1.15</oid> <!-- IF-MIB::ifHighSpeed -->
	</walk>
	<walk correlation-id="3-2" max-repetitions="2">
		<oid>.1.3.6.1.2.1.31.1.1.1.6</oid> <!-- IF-MIB::ifHCInOctets -->
	</walk>
	<walk correlation-id="3-3" max-repetitions="2">
		<oid>.1.3.6.1.2.1.31.1.1.1.10</oid> <!-- IF-MIB::ifHCOutOctets -->
	</walk>
</snmp-request>`

var expectedResponseXML = `<snmp-response>
	<response correlation-id="0">
		<result>
			<base>.1.3.6.1.2.1.2.1</base>
			<instance>.0</instance>
			<value type="2">Aw==</value>
		</result>
	</response>
	<response correlation-id="1">
		<result>
			<base>.1.3.6.1.2.1.1.3</base>
			<instance>.0</instance>
			<value type="67">JxA=</value>
		</result>
	</response>
	<response correlation-id="2-0-0">
		<result>
			<base>.1.3.6.1.2.1.6.5</base>
			<instance>.0</instance>
			<value type="65">Mg==</value>
		</result>
	</response>
	<response correlation-id="2-1-0">
		<result>
			<base>.1.3.6.1.2.1.6.6</base>
			<instance>.0</instance>
			<value type="65">GQ==</value>
		</result>
	</response>
	<response correlation-id="3-0">
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.1</base>
			<instance>.1</instance>
			<value type="4">bDA=</value>
		</result>
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.1</base>
			<instance>.2</instance>
			<value type="4">ZXRoMA==</value>
		</result>
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.1</base>
			<instance>.3</instance>
			<value type="4">d2xhbjA=</value>
		</result>
	</response>
	<response correlation-id="3-1">
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.15</base>
			<instance>.1</instance>
			<value type="66">Cg==</value>
		</result>
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.15</base>
			<instance>.2</instance>
			<value type="66">A+g=</value>
		</result>
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.15</base>
			<instance>.3</instance>
			<value type="66">ASw=</value>
		</result>
	</response>
	<response correlation-id="3-2">
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.6</base>
			<instance>.1</instance>
			<value type="70">AMg=</value>
		</result>
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.6</base>
			<instance>.2</instance>
			<value type="70">AA==</value>
		</result>
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.6</base>
			<instance>.3</instance>
			<value type="70">AZA=</value>
		</result>
	</response>
	<response correlation-id="3-3">
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.10</base>
			<instance>.1</instance>
			<value type="70">ASw=</value>
		</result>
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.10</base>
			<instance>.2</instance>
			<value type="70">AA==</value>
		</result>
		<result>
			<base>.1.3.6.1.2.1.31.1.1.1.10</base>
			<instance>.3</instance>
			<value type="70">Arw=</value>
		</result>
	</response>
</snmp-response>`

func TestSNMPResponse(t *testing.T) {
	req := &api.SNMPRequestDTO{}
	err := xml.Unmarshal([]byte(requestXML), req)
	assert.NilError(t, err)

	client := &tools.MockSNMPClient{
		WalkMap: make(map[string][]gosnmp.SnmpPDU),
	}
	for _, walk := range req.Walks {
		oid := walk.OIDs[0]
		switch oid {
		case ".1.3.6.1.2.1.2.1":
			client.WalkMap[oid] = []gosnmp.SnmpPDU{
				{Name: ".1.3.6.1.2.1.2.1.0", Type: gosnmp.Integer, Value: 3},
			}
		case ".1.3.6.1.2.1.1.3":
			client.WalkMap[oid] = []gosnmp.SnmpPDU{
				{Name: ".1.3.6.1.2.1.1.3.0", Type: gosnmp.TimeTicks, Value: 10000},
			}
		case ".1.3.6.1.2.1.6.5":
			client.WalkMap[oid] = []gosnmp.SnmpPDU{
				{Name: ".1.3.6.1.2.1.6.5.0", Type: gosnmp.Counter32, Value: 50},
			}
		case ".1.3.6.1.2.1.6.6":
			client.WalkMap[oid] = []gosnmp.SnmpPDU{
				{Name: ".1.3.6.1.2.1.6.6.0", Type: gosnmp.Counter32, Value: 25},
			}
		case ".1.3.6.1.2.1.31.1.1.1.1":
			client.WalkMap[oid] = []gosnmp.SnmpPDU{
				{Name: ".1.3.6.1.2.1.31.1.1.1.1.1", Type: gosnmp.OctetString, Value: "l0"},
				{Name: ".1.3.6.1.2.1.31.1.1.1.1.2", Type: gosnmp.OctetString, Value: "eth0"},
				{Name: ".1.3.6.1.2.1.31.1.1.1.1.3", Type: gosnmp.OctetString, Value: "wlan0"},
			}
		case ".1.3.6.1.2.1.31.1.1.1.15":
			client.WalkMap[oid] = []gosnmp.SnmpPDU{
				{Name: ".1.3.6.1.2.1.31.1.1.1.15.1", Type: gosnmp.Gauge32, Value: 10},
				{Name: ".1.3.6.1.2.1.31.1.1.1.15.2", Type: gosnmp.Gauge32, Value: 1000},
				{Name: ".1.3.6.1.2.1.31.1.1.1.15.3", Type: gosnmp.Gauge32, Value: 300},
			}
		case ".1.3.6.1.2.1.31.1.1.1.6":
			client.WalkMap[oid] = []gosnmp.SnmpPDU{
				{Name: ".1.3.6.1.2.1.31.1.1.1.6.1", Type: gosnmp.Counter64, Value: 200},
				{Name: ".1.3.6.1.2.1.31.1.1.1.6.2", Type: gosnmp.Counter64, Value: 0},
				{Name: ".1.3.6.1.2.1.31.1.1.1.6.3", Type: gosnmp.Counter64, Value: 400},
			}
		case ".1.3.6.1.2.1.31.1.1.1.10":
			client.WalkMap[oid] = []gosnmp.SnmpPDU{
				{Name: ".1.3.6.1.2.1.31.1.1.1.10.1", Type: gosnmp.Counter64, Value: 300},
				{Name: ".1.3.6.1.2.1.31.1.1.1.10.2", Type: gosnmp.Counter64, Value: 0},
				{Name: ".1.3.6.1.2.1.31.1.1.1.10.3", Type: gosnmp.Counter64, Value: 700},
			}
		}
	}

	response := snmpModule.getResponse(client, req)
	bytes, err := xml.MarshalIndent(response, "", "	")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
	assert.Equal(t, expectedResponseXML, string(bytes))
	assert.Equal(t, len(req.Walks), len(response.Responses))

	ifName := findResponse(response, "3-0")
	assert.Assert(t, ifName != nil)
	assert.Equal(t, 3, len(ifName.Results))
	assert.Equal(t, ".2", ifName.Results[1].Instance)
	eth0, err := base64.StdEncoding.DecodeString(ifName.Results[1].Value.Value)
	assert.NilError(t, err)
	assert.Equal(t, "eth0", string(eth0))
}

func findResponse(response *api.SNMPMultiResponseDTO, correlationID string) *api.SNMPResponseDTO {
	for _, r := range response.Responses {
		if r.CorrelationID == correlationID {
			return &r
		}
	}
	return nil
}
