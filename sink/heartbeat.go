package sink

import (
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
)

// HeartbeatModule represents the heartbeat module
type HeartbeatModule struct{}

// GetID gets the ID of the sink module
func (module *HeartbeatModule) GetID() string {
	return "Heartbeat"
}

// Start initiates a blocking loop that sends heartbeats to OpenNMS
func (module *HeartbeatModule) Start(config *api.MinionConfig, broker api.Broker) error {
	log.Infof("Starting Sink Heartbeat Module")
	go func() {
		for {
			log.Infof("Sending heartbeat for Minion with id %s at location %s", config.ID, config.Location)
			sendResponse(module.GetID(), config, broker, module.getIdentity(config))
			time.Sleep(30 * time.Second)
		}
	}()
	return nil
}

// Stop shutdowns the sink module
func (module *HeartbeatModule) Stop() {
	log.Warnf("Stopping Sink Heartbeat Module")
}

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
