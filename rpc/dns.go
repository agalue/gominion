package rpc

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
)

// DNSLookupClientRPCModule represents the RPC Module implementation for DNS Lookup
type DNSLookupClientRPCModule struct {
}

// GetID gets the module ID
func (module *DNSLookupClientRPCModule) GetID() string {
	return "DNS"
}

// Execute executes the request synchronously and return the response from the module
func (module *DNSLookupClientRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.DNSLookupRequestDTO{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.DNSLookupResponseDTO{Error: getError(request, err)}
		bytes, _ := xml.Marshal(response)
		return transformResponse(request, bytes)
	}
	response := &api.DNSLookupResponseDTO{}
	if req.QueryType == "LOOKUP" {
		addresses, err := net.LookupIP(req.HostRequest)
		if err != nil || len(addresses) == 0 {
			response.Error = fmt.Sprintf("Cannot lookup for address %s: %v", req.HostRequest, err)
		} else {
			response.HostResponse = addresses[0].String()
		}
	} else if req.QueryType == "REVERSE_LOOKUP" {
		hostnames, err := net.LookupAddr(req.HostRequest)
		if err != nil && len(hostnames) == 0 {
			response.Error = fmt.Sprintf("Cannot reverse lookup for address %s: %v", req.HostRequest, err)
		} else {
			response.HostResponse = hostnames[0]
		}
	} else {
		response.Error = fmt.Sprintf("Invalid query type: %s", req.QueryType)
	}
	log.Printf("Sending DNS %s response for '%s' = '%s'", req.QueryType, req.HostRequest, response.HostResponse)
	bytes, _ := xml.Marshal(response)
	return transformResponse(request, bytes)
}

var dnsModule = &DNSLookupClientRPCModule{}

func init() {
	api.RegisterRPCModule(dnsModule)
}
