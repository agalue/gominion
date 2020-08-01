package sink

import (
	"log"
	"time"

	"github.com/agalue/gominion/api"
)

// HeartbeatModule represents the heartbeat module
type HeartbeatModule struct{}

// GetID gets the ID of the sink module
func (module *HeartbeatModule) GetID() string {
	return "Heartbeat"
}

// Start initiates a blocking loop that sends heartbeats to OpenNMS
func (module *HeartbeatModule) Start(config *api.MinionConfig, broker api.Broker) {
	log.Printf("Starting Sink Heartbeat Module")
	for {
		log.Printf("Sending heartbeat for Minion with id %s at location %s", config.ID, config.Location)
		sendResponse(module.GetID(), config, broker, module.getIdentity(config))
		time.Sleep(30 * time.Second)
	}
}

// Stop shutdowns the sink module
func (module *HeartbeatModule) Stop() {}

func (module *HeartbeatModule) getIdentity(config *api.MinionConfig) *api.MinionIdentityDTO {
	return &api.MinionIdentityDTO{
		ID:        config.ID,
		Location:  config.Location,
		Timestamp: &api.Timestamp{Time: time.Now()},
	}
}

var heartbeatModule = &HeartbeatModule{}

func init() {
	api.RegisterSinkModule(heartbeatModule)
}
