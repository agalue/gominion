package api

import (
	"encoding/xml"
	"testing"

	"gotest.tools/assert"
)

func TestDetectorRequest(t *testing.T) {
	xmlstr := `
	<detector-request location="Test" system-id="minion1" class-name="org.opennms.netmgt.provision.detector.snmp.SnmpDetector" address="192.168.0.1">
		<detector-attribute key="oid">.1.3.6.1.2.1.1.2.0</detector-attribute>
		<runtime-attribute key="read-community">public</runtime-attribute>
		<runtime-attribute key="timeout">3000</runtime-attribute>
	</detector-request>
	`

	// Parse and validate detector request
	request := &DetectorRequestDTO{}
	err := xml.Unmarshal([]byte(xmlstr), request)
	assert.NilError(t, err)
	assert.Equal(t, "Test", request.Location)
	assert.Equal(t, 1, len(request.DetectorAttributes))
	assert.Equal(t, ".1.3.6.1.2.1.1.2.0", request.GetAttributeValue("oid", ""))
	assert.Equal(t, 2, len(request.RuntimeAttributes))
	assert.Equal(t, "public", request.GetRuntimeAttributeValue("read-community"))
	assert.Equal(t, 3000, request.GetRuntimeAttributeValueAsInt("timeout"))
	assert.Equal(t, int64(3000), request.GetTimeout().Milliseconds())
}
