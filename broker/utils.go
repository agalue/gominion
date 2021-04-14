package broker

import (
	"github.com/agalue/gominion/api"
)

// GetBroker gets a broker implementation
func GetBroker(config *api.MinionConfig) api.Broker {
	if config.BrokerType == "grpc" {
		return &GrpcClient{}
	}
	if config.BrokerType == "kafka" {
		return &KafkaClient{}
	}
	return nil
}
