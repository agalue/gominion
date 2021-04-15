package api

import (
	"sync"
)

type SinkRegistry struct {
	sinkRegistryMap   map[string]SinkModule
	sinkRegistryMutex sync.RWMutex
}

// Init initializing the Sink registry
func (r *SinkRegistry) Init() {
	r.sinkRegistryMap = make(map[string]SinkModule)
	r.sinkRegistryMutex = sync.RWMutex{}
}

// RegisterModule registers a new RPC Module implementation
func (r *SinkRegistry) RegisterModule(module SinkModule) {
	r.sinkRegistryMutex.Lock()
	r.sinkRegistryMap[module.GetID()] = module
	r.sinkRegistryMutex.Unlock()
}

// UnregisterModule unregister an existing RPC Module implementation
func (r *SinkRegistry) UnregisterModule(module SinkModule) {
	r.sinkRegistryMutex.Lock()
	delete(r.sinkRegistryMap, module.GetID())
	r.sinkRegistryMutex.Unlock()
}

// GetModule gets the RPC Module implementation for a given ID
func (r *SinkRegistry) GetModule(id string) (SinkModule, bool) {
	r.sinkRegistryMutex.RLock()
	defer r.sinkRegistryMutex.RUnlock()
	module, ok := r.sinkRegistryMap[id]
	return module, ok
}

// GetAllModules gets all the registered Sink modules
func (r *SinkRegistry) GetAllModules() []SinkModule {
	r.sinkRegistryMutex.RLock()
	defer r.sinkRegistryMutex.RUnlock()
	modules := make([]SinkModule, 0, len(r.sinkRegistryMap))
	for _, v := range r.sinkRegistryMap {
		modules = append(modules, v)
	}
	return modules
}
