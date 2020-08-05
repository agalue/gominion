package tools

import (
	"fmt"

	"github.com/soniah/gosnmp"
)

// MockSNMPClient represents a mock implementation of the SNMP Handler for testing purposes
type MockSNMPClient struct {
	WalkMap map[string][]gosnmp.SnmpPDU
	GetMap  map[string]*gosnmp.SnmpPacket
}

// Version returns a fixed version for testing purposes
func (cli *MockSNMPClient) Version() string {
	return "2c"
}

// Target returns a fixed IP for testing purposes
func (cli *MockSNMPClient) Target() string {
	return "127.0.0.1"
}

// Connect always returns a nil error
func (cli *MockSNMPClient) Connect() error {
	return nil
}

// Disconnect always returns a nil error
func (cli *MockSNMPClient) Disconnect() error {
	return nil
}

// BulkWalk emulates a walk based on the provided list of PDUs
func (cli *MockSNMPClient) BulkWalk(rootOid string, walkFn gosnmp.WalkFunc) error {
	if cli.WalkMap == nil || len(cli.WalkMap) == 0 {
		return fmt.Errorf("There was a problem")
	}
	for _, pdu := range cli.WalkMap[rootOid] {
		walkFn(pdu)
	}
	return nil
}

// Get emulates a get based on the provided map of PDU packets
func (cli *MockSNMPClient) Get(oid string) (result *gosnmp.SnmpPacket, err error) {
	if cli.GetMap == nil {
		return nil, fmt.Errorf("There was a problem")
	}
	return cli.GetMap[oid], nil
}
