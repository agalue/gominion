package api

import (
	"log"
	"sync"
)

var sinkRegistryMap map[string]SinkModule = make(map[string]SinkModule)
var sinkRegistryMutex = sync.RWMutex{}

// RegisterSinkModule registers a new RPC Module implementation
func RegisterSinkModule(module SinkModule) {
	log.Printf("Registering Sink module: %s", module.GetID())
	sinkRegistryMutex.Lock()
	sinkRegistryMap[module.GetID()] = module
	sinkRegistryMutex.Unlock()
}

// UnregisterSinkModule unregister an existing RPC Module implementation
func UnregisterSinkModule(module SinkModule) {
	log.Printf("Unregistering Sink module: %s", module.GetID())
	sinkRegistryMutex.Lock()
	delete(sinkRegistryMap, module.GetID())
	sinkRegistryMutex.Unlock()
}

// GetSinkModule gets the RPC Module implementation for a given ID
func GetSinkModule(id string) (SinkModule, bool) {
	sinkRegistryMutex.RLock()
	defer sinkRegistryMutex.RUnlock()
	module, ok := sinkRegistryMap[id]
	return module, ok
}

// GetAllSinkModules gets all the registered Sink modules
func GetAllSinkModules() []SinkModule {
	sinkRegistryMutex.RLock()
	defer sinkRegistryMutex.RUnlock()
	modules := make([]SinkModule, 0, len(sinkRegistryMap))
	for _, v := range sinkRegistryMap {
		modules = append(modules, v)
	}
	return modules
}
