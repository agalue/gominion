package api

import (
	"github.com/agalue/gominion/protobuf/ipc"
)

// SinkModule represents the Sink Module interface
type SinkModule interface {
	GetID() string
	Start(config *MinionConfig, stream ipc.OpenNMSIpc_SinkStreamingClient)
	Stop()
}
