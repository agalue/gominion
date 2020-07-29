package api

import (
	"github.com/agalue/gominion/protobuf/ipc"
)

// RPCModule represents the RPC Module interface
type RPCModule interface {
	GetID() string
	Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto
}
