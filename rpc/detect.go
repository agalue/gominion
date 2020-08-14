package rpc

import (
	"encoding/xml"
	"fmt"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/detectors"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"
)

// DetectorClientRPCModule represents the RPC Module implementation for a detector
type DetectorClientRPCModule struct {
}

// GetID gets the module ID
func (module *DetectorClientRPCModule) GetID() string {
	return "Detect"
}

// Execute executes the detection request synchronously and return the response
func (module *DetectorClientRPCModule) Execute(request *ipc.RpcRequestProto) *ipc.RpcResponseProto {
	req := &api.DetectorRequestDTO{}
	if err := xml.Unmarshal(request.RpcContent, req); err != nil {
		response := &api.DetectorResponseDTO{Error: getError(request, err)}
		return transformResponse(request, response)
	}
	detectorID := req.GetDetector()
	response := &api.DetectorResponseDTO{}
	log.Infof("Executing detector %s against %s", detectorID, req.IPAddress)
	if monitor, ok := detectors.GetDetector(detectorID); ok {
		response = monitor.Detect(req)
	} else {
		response.Error = getError(request, fmt.Errorf("Cannot find implementation for detector %s", detectorID))
	}
	log.Infof("Sending detection status %s on %s as %s", detectorID, req.IPAddress, response.GetStatus())
	return transformResponse(request, response)
}

var detectorModule = &DetectorClientRPCModule{}

func init() {
	api.RegisterRPCModule(detectorModule)
}
