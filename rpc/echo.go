package rpc

import (
	"encoding/xml"
	"log"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
)

// EchoRPCModule represents the RPC Module implementation for Echo
type EchoRPCModule struct {
}

// GetID gets the module ID
func (module *EchoRPCModule) GetID() string {
	return "Echo"
}

// Execute executes the request synchronously and return the response from the module
func (module *EchoRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.EchoRequest{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.EchoResponse{Error: getError(request, err)}
		bytes, _ := xml.Marshal(response)
		return transformResponse(request, bytes)
	}
	if req.Delay > 0 {
		time.Sleep(time.Duration(req.Delay) * time.Microsecond)
	}
	response := &api.EchoResponse{
		ID:      req.ID,
		Message: req.Message,
		Body:    req.Body,
	}
	log.Printf("Sending echo response for ID %v", response.ID)
	bytes, _ := xml.Marshal(response)
	return transformResponse(request, bytes)
}

var echoModule = &EchoRPCModule{}

func init() {
	api.RegisterRPCModule(echoModule)
}
