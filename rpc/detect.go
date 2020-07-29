package rpc

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/detectors"
	"github.com/agalue/gominion/protobuf/ipc"
)

// DetectorClientRPCModule represents the RPC Module implementation for a detector
type DetectorClientRPCModule struct {
}

// GetID gets the module ID
func (module *DetectorClientRPCModule) GetID() string {
	return "Detect"
}

// Execute executes the request synchronously and return the response from the module
func (module *DetectorClientRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.DetectorRequestDTO{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.DetectorResponseDTO{Error: getError(request, err)}
		bytes, _ := xml.Marshal(response)
		return transformResponse(request, bytes)
	}
	detectorID := req.GetDetector()
	response := &api.DetectorResponseDTO{}
	log.Printf("Executing detector %s against %s", detectorID, req.IPAddress)
	if monitor, ok := detectors.GetDetector(detectorID); ok {
		results := monitor.Detect(req)
		response.Detected = results.IsServiceDetected
	} else {
		msg := fmt.Sprintf("Error cannot find implementation for detector %s, ignoring request with ID %s", detectorID, request.RpcId)
		response.Error = msg
		log.Printf(msg)
	}
	log.Printf("Sending detector response for %s on %s ? %v", detectorID, req.IPAddress, response.Detected)
	bytes, _ := xml.Marshal(response)
	return transformResponse(request, bytes)
}

var detectorModule = &DetectorClientRPCModule{}

func init() {
	api.RegisterRPCModule(detectorModule)
}
