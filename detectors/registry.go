package detectors

import (
	"sync"

	"github.com/agalue/gominion/api"
)

var detectorRegistryMap map[string]api.ServiceDetector = make(map[string]api.ServiceDetector)
var detectorRegistryMutex = sync.RWMutex{}

// RegisterDetector registers a new detector implementation
func RegisterDetector(detector api.ServiceDetector) {
	detectorRegistryMutex.Lock()
	detectorRegistryMap[detector.GetID()] = detector
	detectorRegistryMutex.Unlock()
}

// UnregisterDetector unregister an existing detector implementation
func UnregisterDetector(detector api.ServiceDetector) {
	detectorRegistryMutex.Lock()
	delete(detectorRegistryMap, detector.GetID())
	detectorRegistryMutex.Unlock()
}

// GetDetector gets the detector implementation for a given ID
func GetDetector(id string) (api.ServiceDetector, bool) {
	detectorRegistryMutex.RLock()
	defer detectorRegistryMutex.RUnlock()
	detector, ok := detectorRegistryMap[id]
	return detector, ok
}

// GetAllDetectors gets all the registered detector modules
func GetAllDetectors() []api.ServiceDetector {
	detectorRegistryMutex.RLock()
	defer detectorRegistryMutex.RUnlock()
	modules := make([]api.ServiceDetector, 0, len(detectorRegistryMap))
	for _, v := range detectorRegistryMap {
		modules = append(modules, v)
	}
	return modules
}
