package api

import (
	"github.com/agalue/gominion/protobuf/ipc"
)

// Sink represents the broker functionality for sending messages via the Sink API
type Sink interface {

	// Sends a Sink Message to OpenNMS
	Send(msg *ipc.SinkMessage) error
}

// Broker represents a broker implementation to control its life cycle
type Broker interface {

	// Starts the broker (non-blocking method)
	Start() error

	// Shuts down the broker
	Stop()
}

// SinkModule represents an implementation of an OpenNMS Sink Module
type SinkModule interface {

	// Returns the ID of the Sink Module implementation
	GetID() string

	// Starts the Sink Module (non-blocking method)
	Start(config *MinionConfig, sink Sink) error

	// Shuts down the Sink Module (including all listeners)
	Stop()
}

// RPCModule represents an implementation of an OpenNMS RPC Module
type RPCModule interface {

	// Returns the ID of the RPC Module implementation
	GetID() string

	// Executes an RPC request and returns the response
	// The response tells if the operation was successful or not
	Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto
}

// ServiceCollector represents an implementation of a service collector
type ServiceCollector interface {

	// Returns the ID of the Service Collector implementation
	GetID() string

	// Executes the data collection operation from the request
	// The response tells if the operation was successful or not
	Collect(request *CollectorRequestDTO) *CollectorResponseDTO
}

// ServiceDetector represents an implementation of a service detector
type ServiceDetector interface {

	// Returns the ID of the Service Detector implementation
	GetID() string

	// Executes the detection operation from the request
	// The response tells if the operation was successful or not
	Detect(request *DetectorRequestDTO) *DetectorResponseDTO
}

// ServiceMonitor represents an implementation of a service monitor
type ServiceMonitor interface {

	// Returns the ID of the Service Monitor implementation
	GetID() string

	// Executes the polling operation from the request
	// The response tells if the operation was successful or not
	Poll(request *PollerRequestDTO) *PollerResponseDTO
}
