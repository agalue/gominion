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
		pollerResponse := &api.PollerResponseDTO{Error: getError(request, err)}
		bytes, _ := xml.Marshal(pollerResponse)
		return transformResponse(request, bytes)
	}
	pollerResponse := &api.PollerResponseDTO{}
	monitorID := req.GetMonitor()
	log.Printf("Executing monitor %s for service %s through %s", monitorID, req.ServiceName, req.IPAddress)
	if monitor, ok := monitors.GetMonitor(monitorID); ok {
		pollerResponse.Status = monitor.Poll(req)
	} else {
		msg := fmt.Sprintf("Error cannot find implementation for monitor %s, ignoring request with ID %s", monitorID, request.RpcId)
		pollerResponse.Error = msg
		log.Printf(msg)
	}
	log.Printf("Sending polling status of %s on %s as %s", req.ServiceName, req.IPAddress, pollerResponse.Status.StatusName)
	bytes, _ := xml.Marshal(pollerResponse)
	return transformResponse(request, bytes)
}

var pollerModule = &PollerClientRPCModule{}

func init() {
	api.RegisterRPCModule(pollerModule)
}
