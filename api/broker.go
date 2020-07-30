package api

import (
	"github.com/agalue/gominion/protobuf/ipc"
)

// Broker represents a broker implementation
type Broker interface {
	Send(msg *ipc.SinkMessage) error
}
