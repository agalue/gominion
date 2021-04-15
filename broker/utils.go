package broker

import (
	"github.com/agalue/gominion/api"
)

// GetBroker returns a broker implementation
func GetBroker(config *api.MinionConfig, registry *api.SinkRegistry, metrics *api.Metrics) api.Broker {
	if config.BrokerType == "grpc" {
		return &GrpcClient{
			config:   config,
			registry: registry,
			metrics:  metrics,
		}
	}
	if config.BrokerType == "kafka" {
		return &KafkaClient{
			config:   config,
			registry: registry,
			metrics:  metrics,
		}
	}
	return nil
}
