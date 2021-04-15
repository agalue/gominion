package rpc

import (
	"encoding/xml"
	"fmt"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/collectors"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"
)

// CollectorClientRPCModule represents the RPC Module implementation for Data Collection
type CollectorClientRPCModule struct {
}

// GetID gets the module ID
func (module *CollectorClientRPCModule) GetID() string {
	return "Collect"
}

// Execute executes the collection request synchronously and return the response
func (module *CollectorClientRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.CollectorRequestDTO{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.CollectorResponseDTO{Error: getError(request, err)}
		return transformResponse(request, response)
	}
	collectorID := req.GetCollector()
	response := &api.CollectorResponseDTO{}
	log.Infof("Executing %s collector against %s", collectorID, req.CollectionAgent.IPAddress)
	if collector, ok := collectors.GetCollector(collectorID); ok {
		response = collector.Collect(req)
	} else {
		response.Error = getError(request, fmt.Errorf("Cannot find implementation for collector %s", collectorID))
	}
	log.Infof("Sending collection of %s from %s", response.GetStatus(), req.CollectionAgent.IPAddress)
	return transformResponse(request, response)
}

func init() {
	api.RegisterRPCModule(&CollectorClientRPCModule{})
}
