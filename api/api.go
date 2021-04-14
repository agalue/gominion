package api

import (
	"github.com/agalue/gominion/protobuf/ipc"
)

// Sink represents the broker functionality for sending messages to the Sink API
type Sink interface {

	// Sends a Sink Message to OpenNMS
	Send(msg *ipc.SinkMessage) error
}

// Broker represents a broker implementation
type Broker interface {
	// Starts the broker
	Start(config *MinionConfig, metrics *Metrics) error

	// Shutdown the broker
	Stop()
}

// SinkModule represents the Sink Module interface
type SinkModule interface {

	// Returns the ID of the Sink Module implementation
	GetID() string

	// Starts the Sink Module
	Start(config *MinionConfig, sink Sink) error

	// Shutdown the Sink Module
	Stop()
}

// RPCModule represents the RPC Module interface
type RPCModule interface {

	// Returns the ID of the RPC Module implementation
	GetID() string

	// Executes an RPC request, and returns the response
	// The response tells if the operation was successful or not
	Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto
}

// ServiceCollector represents the service collector interface
type ServiceCollector interface {

	// Returns the ID of the Service Collector implementation
	GetID() string

	// Executes the data collection operation from the request
	// The response tells if the operation was successful or not
	Collect(request *CollectorRequestDTO) *CollectorResponseDTO
}

// ServiceDetector represents the service detector interface
type ServiceDetector interface {

	// Returns the ID of the Service Detector implementation
	GetID() string

	// Executes the detection operation from the request
	// The response tells if the operation was successful or not
	Detect(request *DetectorRequestDTO) *DetectorResponseDTO
}

// ServiceMonitor represents the service monitor interface
type ServiceMonitor interface {

	// Returns the ID of the Service Monitor implementation
	GetID() string

	// Executes the polling operation from the request
	// The response tells if the operation was successful or not
	Poll(request *PollerRequestDTO) *PollerResponseDTO
}
