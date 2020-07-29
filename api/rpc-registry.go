package api

import (
	"log"
	"sync"
)

var rpcRegistryMap map[string]RPCModule = make(map[string]RPCModule)
var rpcRegistryMutex = sync.RWMutex{}

// RegisterRPCModule registers a new RPC Module implementation
func RegisterRPCModule(module RPCModule) {
	log.Printf("Registering RPC module: %s", module.GetID())
	rpcRegistryMutex.Lock()
	rpcRegistryMap[module.GetID()] = module
	rpcRegistryMutex.Unlock()
}

// UnregisterRPCModule unregister an existing RPC Module implementation
func UnregisterRPCModule(module RPCModule) {
	log.Printf("Unregistering RPC module: %s", module.GetID())
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
