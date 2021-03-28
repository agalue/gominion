package monitors

import (
	"testing"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/tools"
	"github.com/gosnmp/gosnmp"
	"gotest.tools/v3/assert"
)

func TestMeetCriteria(t *testing.T) {
	assert.Assert(t, !snmpMonitor.meetsCriteria("", "=", "ok"))
	assert.Assert(t, snmpMonitor.meetsCriteria("ok", "", ""))
	assert.Assert(t, snmpMonitor.meetsCriteria("ok", "=", "ok"))
	assert.Assert(t, snmpMonitor.meetsCriteria("ok", "!=", "nok"))
	assert.Assert(t, snmpMonitor.meetsCriteria("3", ">", "2"))
	assert.Assert(t, snmpMonitor.meetsCriteria("3", ">=", "2"))
	assert.Assert(t, snmpMonitor.meetsCriteria("2", "<", "3"))
	assert.Assert(t, snmpMonitor.meetsCriteria("2", "<=", "3"))
	assert.Assert(t, snmpMonitor.meetsCriteria("work", "~", "^w.*k"))
	assert.Assert(t, snmpMonitor.meetsCriteria("work", "~", "\\w+"))
	assert.Assert(t, !snmpMonitor.meetsCriteria("work", "~", "\\d+"))
}

func TestSnmpMonitor(t *testing.T) {
	client := &tools.MockSNMPClient{
		GetMap:  make(map[string]*gosnmp.SnmpPacket),
		WalkMap: make(map[string][]gosnmp.SnmpPDU),
	}
	client.GetMap[defaultOID] = &gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{Value: "ok"},
		},
	}
	client.WalkMap[defaultOID] = []gosnmp.SnmpPDU{
		{Value: "aa"},
		{Value: "bb"},
		{Value: "bb"},
		{Value: "cc"},
		{Value: "dd"},
	}

	// Test Equal
	response := snmpMonitor.poll(client, defaultOID, "false", "false", "=", "ok", 0, 0)
	assert.Equal(t, api.ServiceAvailableCode, response.Status.StatusCode)
	assert.Equal(t, "", response.Error)

	// Test Not Equal
	response = snmpMonitor.poll(client, defaultOID, "false", "false", "!=", "ok", 0, 0)
	assert.Equal(t, api.ServiceUnavailableCode, response.Status.StatusCode)
	assert.Equal(t, "", response.Error)

	// Test Count Equal
	response = snmpMonitor.poll(client, defaultOID, "count", "false", "=", "bb", 2, 2)
	assert.Equal(t, api.ServiceAvailableCode, response.Status.StatusCode)
	assert.Equal(t, "", response.Error)

	// Test Walk
	client.WalkMap[defaultOID] = []gosnmp.SnmpPDU{
		{Value: "battle"},
		{Value: "bottle"},
		{Value: "bath"},
	}
	response = snmpMonitor.poll(client, defaultOID, "false", "true", "~", "^b.*", 0, 0)
	assert.Equal(t, api.ServiceAvailableCode, response.Status.StatusCode)
	assert.Equal(t, "", response.Error)
}
