package broker

import (
	"strings"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/collectors"
	"github.com/agalue/gominion/detectors"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/monitors"

	_ "github.com/agalue/gominion/rpc" // Load all RPC modules
)

// GetBroker returns a broker implementation
func GetBroker(config *api.MinionConfig, registry *api.SinkRegistry, metrics *api.Metrics) api.Broker {
	if strings.ToLower(config.BrokerType) == "grpc" {
		return &GrpcClient{
			config:   config,
			registry: registry,
			metrics:  metrics,
		}
	}
	if strings.ToLower(config.BrokerType) == "kafka" {
		return &KafkaClient{
			config:   config,
			registry: registry,
			metrics:  metrics,
		}
	}
	return nil
}

// DisplayRegisteredModules displays all registered modules
func DisplayRegisteredModules(sinkRegistry *api.SinkRegistry) {
	for _, m := range api.GetAllRPCModules() {
		log.Debugf("Registered RPC module %s", m.GetID())
	}
	for _, m := range sinkRegistry.GetAllModules() {
		log.Debugf("Registered Sink module %s", m.GetID())
	}
	for _, m := range collectors.GetAllCollectors() {
		log.Debugf("Registered collector module %s", m.GetID())
	}
	for _, m := range detectors.GetAllDetectors() {
		log.Debugf("Registered detector module %s", m.GetID())
	}
	for _, m := range monitors.GetAllMonitors() {
		log.Debugf("Registered poller module %s", m.GetID())
	}
}
