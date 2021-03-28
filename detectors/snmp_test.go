package detectors

import (
	"testing"

	"github.com/agalue/gominion/tools"
	"github.com/gosnmp/gosnmp"
	"gotest.tools/v3/assert"
)

func TestSnmpDetectorExist(t *testing.T) {
	client := &tools.MockSNMPClient{
		GetMap: make(map[string]*gosnmp.SnmpPacket),
	}
	client.GetMap[defaultOID] = &gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{Value: "ok"},
		},
	}
	response := snmpDetector.detect(client, defaultOID, "Exist", "false", "")
	assert.Equal(t, true, response.Detected)
	assert.Equal(t, "", response.Error)
}

func TestSnmpDetectorAny(t *testing.T) {
	client := &tools.MockSNMPClient{
		WalkMap: make(map[string][]gosnmp.SnmpPDU),
	}
	client.WalkMap[defaultOID] = []gosnmp.SnmpPDU{
		{Value: "a"},
		{Value: "b"},
		{Value: "c"},
	}

	// b is included in the responses
	response := snmpDetector.detect(client, defaultOID, "Any", "true", "b")
	assert.Equal(t, true, response.Detected)
	assert.Equal(t, "", response.Error)

	// d is not included in the responses
	response = snmpDetector.detect(client, defaultOID, "Any", "true", "d")
	assert.Equal(t, false, response.Detected)
	assert.Equal(t, "", response.Error)

	// expected value is required
	response = snmpDetector.detect(client, defaultOID, "Any", "true", "")
	assert.Equal(t, false, response.Detected)
	assert.Assert(t, response.Error != "")
}

func TestSnmpDetectorNone(t *testing.T) {
	client := &tools.MockSNMPClient{
		WalkMap: make(map[string][]gosnmp.SnmpPDU),
	}
	client.WalkMap[defaultOID] = []gosnmp.SnmpPDU{
		{Value: "a"},
		{Value: "b"},
		{Value: "c"},
	}

	// b is included in the responses
	response := snmpDetector.detect(client, defaultOID, "None", "true", "b")
	assert.Equal(t, false, response.Detected)
	assert.Equal(t, "", response.Error)

	// d is not included in the responses
	response = snmpDetector.detect(client, defaultOID, "None", "true", "d")
	assert.Equal(t, true, response.Detected)
	assert.Equal(t, "", response.Error)

	// expected value is required
	response = snmpDetector.detect(client, defaultOID, "None", "true", "")
	assert.Equal(t, false, response.Detected)
	assert.Assert(t, response.Error != "")
}

func TestSnmpDetectorAll(t *testing.T) {
	client := &tools.MockSNMPClient{
		WalkMap: make(map[string][]gosnmp.SnmpPDU),
	}
	client.WalkMap[defaultOID] = []gosnmp.SnmpPDU{
		{Value: "aa"},
		{Value: "aa"},
		{Value: "aa"},
	}

	// aa is included in all the responses
	response := snmpDetector.detect(client, defaultOID, "All", "true", "aa")
	assert.Equal(t, true, response.Detected)
	assert.Equal(t, "", response.Error)

	client.WalkMap[defaultOID] = []gosnmp.SnmpPDU{
		{Value: "aa"},
		{Value: "bb"},
		{Value: "aa"},
	}

	// aa is not included in all the responses
	response = snmpDetector.detect(client, defaultOID, "All", "true", "aa")
	assert.Equal(t, false, response.Detected)
	assert.Equal(t, "", response.Error)

	// expected value is required
	response = snmpDetector.detect(client, defaultOID, "All", "true", "")
	assert.Equal(t, false, response.Detected)
	assert.Assert(t, response.Error != "")
}
