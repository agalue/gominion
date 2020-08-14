package rpc

import (
	"encoding/xml"
	"fmt"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
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

// Execute executes the ping request synchronously and return the response
func (module *PingProxyRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.PingRequest{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.PingResponse{Error: getError(request, err)}
		return transformResponse(request, response)
	}
	response := &api.PingResponse{}
	if duration, err := tools.Ping(req.Address, req.GetTimeout()); err == nil {
		response.RTT = duration.Seconds()
	} else {
		response.Error = getError(request, fmt.Errorf("Cannot ping address %s: %v", req.Address, err))
	}
	log.Debugf("Sending Ping response for %s", req.Address)
	return transformResponse(request, response)
}

var pingProxyModule = &PingProxyRPCModule{}

func init() {
	api.RegisterRPCModule(pingProxyModule)
}
