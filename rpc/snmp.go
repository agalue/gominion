package rpc

import (
	"encoding/xml"
	"log"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/agalue/gominion/tools"
	"github.com/soniah/gosnmp"
)

// SNMPProxyRPCModule represents the RPC Module implementation for SNMP
type SNMPProxyRPCModule struct {
}

// GetID gets the module ID
func (module *SNMPProxyRPCModule) GetID() string {
	return "SNMP"
}

// Execute executes the request synchronously and return the response from the module
func (module *SNMPProxyRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.SNMPRequestDTO{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.SNMPMultiResponseDTO{Error: getError(request, err)}
		return transformResponse(request, response)
	}
	client := req.Agent.GetSNMPClient()
	if err := client.Connect(); err != nil {
		response := &api.SNMPMultiResponseDTO{Error: getError(request, err)}
		return transformResponse(request, response)
	}
	defer client.Conn.Close()
	response := &api.SNMPMultiResponseDTO{}
	for _, walk := range req.Walks {
		response.AddResponse(module.snmpWalk(client, walk))
	}
	return transformResponse(request, response)
}

func (module *SNMPProxyRPCModule) snmpWalk(client *gosnmp.GoSNMP, walk api.SNMPWalkRequestDTO) *api.SNMPResponseDTO {
	response := &api.SNMPResponseDTO{CorrelationID: walk.CorrelationID}
	log.Printf("Executing %d snmpwalk %s against %s", len(walk.OIDs), client.Version.String(), client.Target)
	for _, oid := range walk.OIDs {
		effectiveOid := tools.GetOidToWalk(oid, walk.Instance)
		err := client.BulkWalk(effectiveOid, func(pdu gosnmp.SnmpPDU) error {
			response.Results = append(response.Results, tools.GetResultForPDU(pdu, oid))
			return nil
		})
		if err != nil {
			log.Printf("Error while walking %s: %v\n", effectiveOid, err)
			return response
		}
	}
	log.Printf("Sending %d snmpwalk responses from %s", len(response.Results), client.Target)
	return response
}

var snmpModule = &SNMPProxyRPCModule{}

func init() {
	api.RegisterRPCModule(snmpModule)
}
