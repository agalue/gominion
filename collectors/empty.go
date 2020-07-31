package collectors

// Placeholder for implementing new collectors

import (
	"time"

	"github.com/agalue/gominion/api"
)

// EmptyCollector represents a collector implementation
type EmptyCollector struct {
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (collector *EmptyCollector) GetID() string {
	return "EMPTY"
}

// Collect execute the collector request and return the collection set
func (collector *EmptyCollector) Collect(request *api.CollectorRequestDTO) api.CollectorResponseDTO {
	response := api.CollectorResponseDTO{
		Error: "Not Implemented",
		CollectionSet: &api.CollectionSetDTO{
			Timestamp: &api.Timestamp{Time: time.Now()},
			Status:    api.CollectionStatusUnknown,
			Agent:     request.CollectionAgent,
		},
	}
	return response
}

var emptyCollector = &EmptyCollector{}

func init() {
	//	RegisterCollector(emptyCollector)
}
