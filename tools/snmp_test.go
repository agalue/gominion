package tools

import (
	"encoding/base64"
	"testing"

	"github.com/agalue/gominion/log"
	"github.com/gosnmp/gosnmp"
	"gotest.tools/v3/assert"
)

func TestGetResultForPDU(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestGetResultForPDU")
	}
	log.InitLogger("debug")
	gosnmp.Default.Target = "127.0.0.1"
	err := gosnmp.Default.Connect()
	assert.NilError(t, err)
	defer gosnmp.Default.Conn.Close()
	oids := []string{
		".1.3.6.1.2.1.1",
		".1.3.6.1.2.1.2.2",
		".1.3.6.1.2.1.31.1.1",
		".1.3.6.1.2.1.4.20",
		".1.3.6.1.2.1.4.34",
	}
	for _, oid := range oids {
		err = gosnmp.Default.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
			result := GetResultForPDU(pdu, oid)
			value, err := base64.StdEncoding.DecodeString(result.Value.Value)
			if err == nil {
				switch result.Value.Type {
				case 6: // ObjectIdentifier
				case 4: // OctetString
					log.Infof("Instance: %s, Type: %d, Value %s", result.Instance, result.Value.Type, string(value))
				default:
					log.Infof("Instance: %s, Type: %d, Value %s", result.Instance, result.Value.Type, gosnmp.ToBigInt(value).String())
				}
			}
			return err
		})
		assert.NilError(t, err)
	}
}
