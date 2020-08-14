package rpc

import (
	"encoding/xml"
	"fmt"
	"net"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"
)

// DNSLookupClientRPCModule represents the RPC Module implementation for DNS Lookup
type DNSLookupClientRPCModule struct {
}

// GetID gets the module ID
func (module *DNSLookupClientRPCModule) GetID() string {
	return "DNS"
}

// Execute executes the DNS request synchronously and return the response
func (module *DNSLookupClientRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.DNSLookupRequestDTO{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.DNSLookupResponseDTO{Error: getError(request, err)}
		return transformResponse(request, response)
	}
	response := &api.DNSLookupResponseDTO{}
	if req.QueryType == "LOOKUP" {
		addresses, err := net.LookupIP(req.HostRequest)
		if err != nil || len(addresses) == 0 {
			response.Error = getError(request, fmt.Errorf("Cannot lookup for address %s: %v", req.HostRequest, err))
		} else {
			response.HostResponse = addresses[0].String()
		}
	} else if req.QueryType == "REVERSE_LOOKUP" {
		hostnames, err := net.LookupAddr(req.HostRequest)
		if err != nil && len(hostnames) == 0 {
			response.Error = getError(request, fmt.Errorf("Cannot reverse lookup for address %s: %v", req.HostRequest, err))
		} else {
			response.HostResponse = hostnames[0]
		}
	} else {
		response.Error = getError(request, fmt.Errorf("Invalid query type: %s", req.QueryType))
	}
	log.Debugf("Sending DNS %s response for %s as %s", req.QueryType, req.HostRequest, response.HostResponse)
	return transformResponse(request, response)
}

var dnsModule = &DNSLookupClientRPCModule{}

func init() {
	api.RegisterRPCModule(dnsModule)
}
