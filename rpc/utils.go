package rpc

import (
	"encoding/xml"
	"fmt"

	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"
)

func transformResponse(request *ipc.RpcRequestProto, object interface{}) *ipc.RpcResponseProto {
	bytes, err := xml.MarshalIndent(object, "", "   ")
	if err != nil {
		log.Errorf("Cannot parse RPC content: %v", err)
		return nil
	}
	response := &ipc.RpcResponseProto{
		ModuleId:   request.ModuleId,
		Location:   request.Location,
		SystemId:   request.SystemId,
		RpcId:      request.RpcId,
		RpcContent: bytes,
	}
	return response
}

func getError(request *ipc.RpcRequestProto, err error) string {
	msg := fmt.Sprintf("Cannot process %s RPC Request with ID %s: %v", request.ModuleId, request.RpcId, err)
	log.Debugf(msg)
	log.Debugf(string(request.RpcContent))
	return msg
}
