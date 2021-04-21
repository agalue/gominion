package collectors

import (
	"fmt"
	"runtime"

	"github.com/agalue/gominion/api"
)

// JMXCollector represents a collector implementation
type JMXCollector struct {
}

// GetID gets the collector ID (simple class name from its Java counterpart)
func (collector *JMXCollector) GetID() string {
	return "Jsr160Collector"
}

// Collect execute the JMX collector request and return the collection response.
// Currently returns mock data for the JMX-Minion service.
func (collector *JMXCollector) Collect(request *api.CollectorRequestDTO) *api.CollectorResponseDTO {
	response := new(api.CollectorResponseDTO)
	agent := request.CollectionAgent
	if agent.IPAddress == "127.0.0.1" && agent.ForeignID == request.SystemID {
		// Mock content for JMX-Minion
		builder := api.NewCollectionSetBuilder(request.CollectionAgent)
		node := &api.CollectionResourceDTO{
			ResourceType: &api.NodeLevelResourceDTO{
				NodeID: request.CollectionAgent.NodeID,
			},
		}
		for _, attr := range collector.getAttributes(request) {
			builder.WithMetric(node, attr)
		}
		response.CollectionSet = builder.Build()
	} else {
		response.MarkAsFailed(request.CollectionAgent, fmt.Errorf("not implemented"))
	}
	return response
}

// Mock content for JMX-Minion
func (collector *JMXCollector) getAttributes(request *api.CollectorRequestDTO) []api.ResourceAttributeDTO {
	attributes := make([]api.ResourceAttributeDTO, 2)
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	attributes[0] = api.ResourceAttributeDTO{
		Name:       "TotalMemory",
		Group:      "java_lang_type_OperatingSystem",
		Identifier: "JMX_java.lang:type=OperatingSystem.TotalMemory",
		Type:       "gauge",
		Value:      fmt.Sprintf("%d", stats.Sys),
	}
	attributes[1] = api.ResourceAttributeDTO{
		Name:       "ThreadCount",
		Group:      "java_lang_type_Threading",
		Identifier: "JMX_java.lang:type=Threading.ThreadCount",
		Type:       "gauge",
		Value:      fmt.Sprintf("%d", runtime.NumGoroutine()),
	}
	return attributes
}

func init() {
	RegisterCollector(&JMXCollector{})
}
