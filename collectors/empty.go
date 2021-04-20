package collectors

// Placeholder for implementing new collectors

import (
	"fmt"

	"github.com/agalue/gominion/api"
)

// EmptyCollector represents a collector implementation
type EmptyCollector struct {
}

// GetID gets the collector ID (simple class name from its Java counterpart)
func (collector *EmptyCollector) GetID() string {
	return "XXX"
}

// Collect execute the XXX collector request and return the collection response
func (collector *EmptyCollector) Collect(request *api.CollectorRequestDTO) *api.CollectorResponseDTO {
	response := new(api.CollectorResponseDTO)
	response.MarkAsFailed(request.CollectionAgent, fmt.Errorf("not implemented"))
	return response
}

func init() {
	//	RegisterCollector(&EmptyCollector{})
}
