package rpc

import (
	"encoding/xml"
	"fmt"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/monitors"
	"github.com/agalue/gominion/protobuf/ipc"
)

// PollerClientRPCModule represents the RPC Module implementation for the Poller client
type PollerClientRPCModule struct {
}

// GetID gets the module ID
func (module *PollerClientRPCModule) GetID() string {
	return "Poller"
}

// Execute executes the polling request synchronously and return the response
func (module *PollerClientRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.PollerRequestDTO{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.PollerResponseDTO{Error: getError(request, err)}
		return transformResponse(request, response)
	}
	response := &api.PollerResponseDTO{}
	monitorID := req.GetMonitor()
	log.Debugf("Executing monitor %s for service %s through %s", monitorID, req.ServiceName, req.IPAddress)
	if monitor, ok := monitors.GetMonitor(monitorID); ok {
		response = monitor.Poll(req)
	} else {
		response.Error = getError(request, fmt.Errorf("cannot find implementation for monitor %s", monitorID))
	}
	log.Debugf("Sending polling status of %s on %s as %s", req.ServiceName, req.IPAddress, response.Status.StatusName)
	return transformResponse(request, response)
}

func init() {
	api.RegisterRPCModule(&PollerClientRPCModule{})
}
