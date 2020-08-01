package sink

import (
	"encoding/xml"
	"log"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/google/uuid"
)

func sendResponse(moduleID string, config *api.MinionConfig, broker api.Broker, object interface{}) {
	bytes, err := xml.MarshalIndent(object, "", "   ")
	if err != nil {
		log.Printf("Error cannot parse Sink API response: %v", err)
	}
	msg := &ipc.SinkMessage{
		MessageId: uuid.New().String(),
		ModuleId:  moduleID,
		SystemId:  config.ID,
		Location:  config.Location,
		Content:   bytes,
	}
	if broker == nil {
		return
	}
	if err := broker.Send(msg); err != nil {
		log.Printf("Error while sending %s message via Sink API: %v", moduleID, err)
	}
}
