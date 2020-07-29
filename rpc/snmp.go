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
		bytes, _ := xml.Marshal(response)
		return transformResponse(request, bytes)
	}
	client := req.Agent.GetSNMPClient()
	if err := client.Connect(); err != nil {
		response := &api.SNMPMultiResponseDTO{Error: getError(request, err)}
		bytes, _ := xml.Marshal(response)
		return transformResponse(request, bytes)
	}
	defer client.Conn.Close()
	multiResponse := &api.SNMPMultiResponseDTO{}
	for _, walk := range req.Walks {
		response := snmpModule.snmpWalk(client, walk)
		multiResponse.Responses = append(multiResponse.Responses, response)
	}
	bytes, err := xml.Marshal(multiResponse)
	if err == nil {
		return transformResponse(request, bytes)
	}
	response := &api.SNMPMultiResponseDTO{Error: getError(request, err)}
	bytes, _ = xml.Marshal(response)
	return transformResponse(request, bytes)
}

func (module *SNMPProxyRPCModule) snmpWalk(client *gosnmp.GoSNMP, walk api.SNMPWalkRequestDTO) api.SNMPResponseDTO {
	response := api.SNMPResponseDTO{CorrelationID: walk.CorrelationID}
	for _, oid := range walk.OIDs {
		effectiveOid := tools.GetOidToWalk(oid, walk.Instance)
		log.Printf("Executing snmpwalk %s for %s against %s", client.Version.String(), effectiveOid, client.Target)
		err := client.BulkWalk(effectiveOid, func(pdu gosnmp.SnmpPDU) error {
			response.Results = append(response.Results, tools.GetResultForPDU(pdu, oid))
			return nil
		})
		if err != nil {
			log.Printf("Error while walking %s: %v\n", effectiveOid, err)
			return response
		}
	}
	return response
}

var snmpModule = &SNMPProxyRPCModule{}

func init() {
	api.RegisterRPCModule(snmpModule)
}
