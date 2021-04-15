package rpc

import (
	"encoding/xml"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"
)

// EchoRPCModule represents the RPC Module implementation for Echo
type EchoRPCModule struct {
}

// GetID gets the module ID
func (module *EchoRPCModule) GetID() string {
	return "Echo"
}

// Execute executes the echo request synchronously and return the response
func (module *EchoRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.EchoRequest{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.EchoResponse{Error: getError(request, err)}
		return transformResponse(request, response)
	}
	if req.Delay > 0 {
		time.Sleep(time.Duration(req.Delay) * time.Microsecond)
	}
	response := &api.EchoResponse{
		ID:      req.ID,
		Message: req.Message,
		Body:    req.Body,
	}
	log.Debugf("Sending echo response for ID %v", response.ID)
	return transformResponse(request, response)
}

func init() {
	api.RegisterRPCModule(&EchoRPCModule{})
}
