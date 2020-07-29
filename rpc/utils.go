package rpc

import (
	"fmt"
	"log"

	"github.com/agalue/gominion/protobuf/ipc"
)

func transformResponse(request *ipc.RpcRequestProto, content []byte) *ipc.RpcResponseProto {
	response := &ipc.RpcResponseProto{
		ModuleId:   request.ModuleId,
		Location:   request.Location,
		SystemId:   request.SystemId,
		RpcId:      request.RpcId,
		RpcContent: content,
	}
	return response
}

func getError(request *ipc.RpcRequestProto, err error) string {
	msg := fmt.Sprintf("Error while parsing %s RPC Request with ID %s: %v", request.ModuleId, request.RpcId, err)
	log.Printf("%s\n%s", msg, string(request.RpcContent))
	return msg
}
