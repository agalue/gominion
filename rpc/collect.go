package rpc

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/collectors"
	"github.com/agalue/gominion/protobuf/ipc"
)

// CollectorClientRPCModule represents the RPC Module implementation for Data Collection
type CollectorClientRPCModule struct {
}

// GetID gets the module ID
func (module *CollectorClientRPCModule) GetID() string {
	return "Collect"
}

// Execute executes the request synchronously and return the response from the module
func (module *CollectorClientRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.CollectorRequestDTO{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.CollectorResponseDTO{Error: getError(request, err)}
		bytes, _ := xml.Marshal(response)
		return transformResponse(request, bytes)
	}
	collectorID := req.GetCollector()
	response := api.CollectorResponseDTO{}
	if collector, ok := collectors.GetCollector(collectorID); ok {
		response = collector.Collect(req)
	} else {
		msg := fmt.Sprintf("Error cannot find implementation for collector %s, ignoring request with ID %s", collectorID, request.RpcId)
		response.Error = msg
		log.Printf(msg)
	}
	bytes, _ := xml.Marshal(response)
	return transformResponse(request, bytes)
}

var collectModule = &CollectorClientRPCModule{}

func init() {
	api.RegisterRPCModule(collectModule)
}
