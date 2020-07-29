package rpc

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/agalue/gominion/tools"
)

// PingProxyRPCModule represents the RPC Module implementation for ICMP pings
type PingProxyRPCModule struct {
}

// GetID gets the module ID
func (module *PingProxyRPCModule) GetID() string {
	return "PING"
}

// Execute executes the request synchronously and return the response from the module
func (module *PingProxyRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.PingRequest{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.PingResponse{Error: getError(request, err)}
		bytes, _ := xml.Marshal(response)
		return transformResponse(request, bytes)
	}
	response := &api.PingResponse{}
	if duration, err := tools.Ping(req.Address, req.GetTimeout()); err == nil {
		response.RTT = duration.Seconds()
	} else {
		response.Error = fmt.Sprintf("Cannot ping address %s: %v", req.Address, err)
	}
	log.Printf("Sending Ping response for %s", req.Address)
	bytes, _ := xml.Marshal(response)
	return transformResponse(request, bytes)
}

var pingProxyModule = &PingProxyRPCModule{}

func init() {
	api.RegisterRPCModule(pingProxyModule)
}
