package collectors

import (
	"log"
	"sync"

	"github.com/agalue/gominion/api"
)

var collectorRegistryMap map[string]api.ServiceCollector = make(map[string]api.ServiceCollector)
var collectorRegistryMutex = sync.RWMutex{}

// RegisterCollector registers a new collector implementation
func RegisterCollector(collector api.ServiceCollector) {
	log.Printf("Registering collector: %s", collector.GetID())
	collectorRegistryMutex.Lock()
	collectorRegistryMap[collector.GetID()] = collector
	collectorRegistryMutex.Unlock()
}

// UnregisterCollector unregister an existing collector implementation
func UnregisterCollector(collector api.ServiceCollector) {
	log.Printf("Unregistering collector: %s", collector.GetID())
	collectorRegistryMutex.Lock()
	delete(collectorRegistryMap, collector.GetID())
	collectorRegistryMutex.Unlock()
}

// GetCollector gets the collector implementation for a given ID
func GetCollector(id string) (api.ServiceCollector, bool) {
	collectorRegistryMutex.RLock()
	defer collectorRegistryMutex.RUnlock()
	collector, ok := collectorRegistryMap[id]
	return collector, ok
}
