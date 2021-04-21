package api

import (
	"fmt"
)

// SinkRegistry track all the enabled Sink module instances.
type SinkRegistry struct {
	sinkRegistryMap map[string]SinkModule
}

// Init initializing the Sink registry
func (r *SinkRegistry) Init() {
	r.sinkRegistryMap = make(map[string]SinkModule)
}

// RegisterModule registers a new RPC Module implementation
func (r *SinkRegistry) RegisterModule(module SinkModule) {
	r.sinkRegistryMap[module.GetID()] = module
}

// UnregisterModule unregister an existing RPC Module implementation
func (r *SinkRegistry) UnregisterModule(module SinkModule) {
	delete(r.sinkRegistryMap, module.GetID())
}

// GetAllModules gets all the registered Sink modules
func (r *SinkRegistry) GetAllModules() []SinkModule {
	modules := make([]SinkModule, 0, len(r.sinkRegistryMap))
	for _, v := range r.sinkRegistryMap {
		modules = append(modules, v)
	}
	return modules
}

// StartModules starts all the registered Sink modules
func (r *SinkRegistry) StartModules(config *MinionConfig, sink Sink) error {
	for _, m := range r.sinkRegistryMap {
		if err := m.Start(config, sink); err != nil {
			return fmt.Errorf("cannot start Sink API module %s: %v", m.GetID(), err)
		}
	}
	return nil
}

// StopModules stops all the registered Sink modules
func (r *SinkRegistry) StopModules() {
	for _, m := range r.sinkRegistryMap {
		m.Stop()
	}
}
