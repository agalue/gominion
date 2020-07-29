package api

import (
	"encoding/xml"
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestDetectorRequest(t *testing.T) {
	request := &DetectorRequestDTO{
		Location:  "Test",
		SystemID:  "minion1",
		ClassName: "org.opennms.netmgt.provision.detector.snmp.SnmpDetector",
		IPAddress: "192.168.0.1",
		DetectorAttributes: []DetectorAttributeDTO{
			{
				Key:   "oid",
				Value: ".1.3.6.1.2.1.1.2.0",
			},
		},
		RuntimeAttributes: []DetectorAttributeDTO{
			{
				Key:   "read-community",
				Value: "public",
			},
			{
				Key:   "timeout",
				Value: "3000",
			},
		},
	}
	bytes, err := xml.MarshalIndent(request, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
	assert.Equal(t, "SnmpDetector", request.GetDetector())
}
