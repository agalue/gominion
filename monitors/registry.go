package monitors

import (
	"log"
	"sync"

	"github.com/agalue/gominion/api"
)

var monitorRegistryMap map[string]api.ServiceMonitor = make(map[string]api.ServiceMonitor)
var monitorRegistryMutex = sync.RWMutex{}

// RegisterMonitor registers a new poller monitor implementation
func RegisterMonitor(monitor api.ServiceMonitor) {
	log.Printf("Registering poller monitor: %s", monitor.GetID())
	monitorRegistryMutex.Lock()
	monitorRegistryMap[monitor.GetID()] = monitor
	monitorRegistryMutex.Unlock()
}

// UnregisterMonitor unregister an existing polller monitor implementation
func UnregisterMonitor(monitor api.ServiceMonitor) {
	log.Printf("Unregistering poller monitor: %s", monitor.GetID())
	monitorRegistryMutex.Lock()
	delete(monitorRegistryMap, monitor.GetID())
	monitorRegistryMutex.Unlock()
}

// GetMonitor gets the poller monitor implementation for a given ID
func GetMonitor(id string) (api.ServiceMonitor, bool) {
	monitorRegistryMutex.RLock()
	defer monitorRegistryMutex.RUnlock()
	monitor, ok := monitorRegistryMap[id]
	return monitor, ok
}
