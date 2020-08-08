package api

import (
	"encoding/xml"
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestCollectionSetBuilder(t *testing.T) {
	builder := NewCollectionSetBuilder(&CollectionAgentDTO{NodeID: 1, IPAddress: "10.0.0.1"})
	nodeType := &NodeLevelResourceDTO{NodeID: 1, Path: "snmp/1"}
	node := &CollectionResourceDTO{ResourceType: nodeType}
	builder.WithAttribute(node, "mib2-tcp", "tcpActiveOpens", "Counter32", "100")
	builder.WithAttribute(node, "mib2-tcp", "tcpPassiveOpens", "Counter32", "200")
	eth0 := &CollectionResourceDTO{
		Name: "eth0",
		ResourceType: &InterfaceLevelResourceDTO{
			Node:     nodeType,
			IntfName: "eth0",
		},
	}
	builder.WithAttribute(eth0, "mib2-X-interfaces", "ifHighSpeed", "string", "1000")
	builder.WithAttribute(eth0, "mib2-X-interfaces", "ifHCInOctets", "Counter64", "100")
	builder.WithAttribute(eth0, "mib2-X-interfaces", "ifHCOutOctets", "Counter64", "200")
	fsRoot := &CollectionResourceDTO{
		Name: "/",
		ResourceType: &GenericTypeResourceDTO{
			Node:     nodeType,
			Instance: "_root_fs",
			Name:     "hrStorageIndex",
		},
	}
	builder.WithAttribute(fsRoot, "mib2-host-resources-storage", "hrStorageUsed", "gauge", "1000")
	cs := builder.Build()
	bytes, err := xml.MarshalIndent(cs, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
	assert.Equal(t, 3, len(cs.Resources))
}
