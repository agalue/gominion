package sink

import (
	"encoding/xml"
	"log"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/google/uuid"
)

// HeartbeatModule represents the heartbeat module
type HeartbeatModule struct{}

// GetID gets the ID of the sink module
func (module *HeartbeatModule) GetID() string {
	return "Heartbeat"
}

// Start initiates a blocking loop that sends heartbeats to OpenNMS
func (module *HeartbeatModule) Start(config *api.MinionConfig, stream ipc.OpenNMSIpc_SinkStreamingClient) {
	log.Printf("Starting Sink Heartbeat Module")
	for {
		log.Printf("Sending heartbeat for Minion with id %s at location %s", config.ID, config.Location)
		msg := module.getSinkMessage(config)
		if err := stream.Send(msg); err != nil {
			log.Printf("Error while sending heartbeat: %v", err)
		}
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

func (module *HeartbeatModule) getSinkMessage(config *api.MinionConfig) *ipc.SinkMessage {
	identity := module.getIdentity(config)
	bytes, _ := xml.Marshal(identity)
	return &ipc.SinkMessage{
		MessageId: uuid.New().String(),
		ModuleId:  "Heartbeat",
		SystemId:  config.ID,
		Location:  config.Location,
		Content:   bytes,
	}
}

var heartbeatModule = &HeartbeatModule{}

func init() {
	api.RegisterSinkModule(heartbeatModule)
}
