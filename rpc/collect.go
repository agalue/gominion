package rpc

import (
	"encoding/xml"
	"log"

	"github.com/agalue/gominion/api"
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
	// FIXME: Start
	response := &api.CollectorResponseDTO{Error: "Implementation comming soon"}
	log.Printf("%s", string(request.RpcContent))
	log.Printf("Ignoring executing of collector %s against %s", collectorID, req.CollectionAgent.IPAddress)
	// FIXME: End
	bytes, _ := xml.Marshal(response)
	return transformResponse(request, bytes)
}

var collectModule = &CollectorClientRPCModule{}

func init() {
	api.RegisterRPCModule(collectModule)
}
