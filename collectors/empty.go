package collectors

// Placeholder for implementing new collectors

import (
	"github.com/agalue/gominion/api"
)

// EmptyCollector represents a collector implementation
type EmptyCollector struct {
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (detector *EmptyCollector) GetID() string {
	return "EMPTY"
}

// Collect execute the collector request and return the collection set
func (detector *EmptyCollector) Collect(request *api.CollectorRequestDTO) api.CollectionSetDTO {
	results := api.CollectionSetDTO{
		Status: api.CollectionStatusUnknown,
	}
	return results
}

var emptyCollector = &EmptyCollector{}

func init() {
	//	RegisterDetector(emptyCollector)
}
