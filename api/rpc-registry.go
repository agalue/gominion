package api

import (
	"sync"
)

var rpcRegistryMap map[string]RPCModule = make(map[string]RPCModule)
var rpcRegistryMutex = sync.RWMutex{}

// RegisterRPCModule registers a new RPC Module implementation
func RegisterRPCModule(module RPCModule) {
	rpcRegistryMutex.Lock()
	rpcRegistryMap[module.GetID()] = module
	rpcRegistryMutex.Unlock()
}

// UnregisterRPCModule unregister an existing RPC Module implementation
func UnregisterRPCModule(module RPCModule) {
	rpcRegistryMutex.Lock()
	delete(rpcRegistryMap, module.GetID())
	rpcRegistryMutex.Unlock()
}

// GetRPCModule gets the RPC Module implementation for a given ID
func GetRPCModule(id string) (RPCModule, bool) {
	rpcRegistryMutex.RLock()
	defer rpcRegistryMutex.RUnlock()
	module, ok := rpcRegistryMap[id]
	return module, ok
}

// GetAllRPCModules gets all the registered RPC modules
func GetAllRPCModules() []RPCModule {
	rpcRegistryMutex.RLock()
	defer rpcRegistryMutex.RUnlock()
	modules := make([]RPCModule, 0, len(rpcRegistryMap))
	for _, v := range rpcRegistryMap {
		modules = append(modules, v)
	}
	return modules
}
