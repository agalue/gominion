package rpc

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/agalue/gominion/api"
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

// Execute executes the request synchronously and return the response from the module
func (module *PollerClientRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.PollerRequestDTO{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.PollerResponseDTO{Error: getError(request, err)}
		return transformResponse(request, response)
	}
	response := &api.PollerResponseDTO{}
	monitorID := req.GetMonitor()
	log.Printf("Executing monitor %s for service %s through %s", monitorID, req.ServiceName, req.IPAddress)
	if monitor, ok := monitors.GetMonitor(monitorID); ok {
		response = monitor.Poll(req)
	} else {
		response.Error = getError(request, fmt.Errorf("Cannot find implementation for monitor %s", monitorID))
	}
	log.Printf("Sending polling status of %s on %s as %s", req.ServiceName, req.IPAddress, response.Status.StatusName)
	return transformResponse(request, response)
}

var pollerModule = &PollerClientRPCModule{}

func init() {
	api.RegisterRPCModule(pollerModule)
}
