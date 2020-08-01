package sink

import (
	"encoding/xml"
	"log"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/agalue/gominion/protobuf/telemetry"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
)

func sendResponse(moduleID string, config *api.MinionConfig, broker api.Broker, object interface{}) {
	bytes, err := xml.MarshalIndent(object, "", "   ")
	if err != nil {
		log.Printf("Error cannot parse Sink API response: %v", err)
	}
	sendBytes(moduleID, config, broker, bytes)
}

func sendBytes(moduleID string, config *api.MinionConfig, broker api.Broker, bytes []byte) {
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

func wrapMessageToTelemetry(config *api.MinionConfig, sourceAddress string, data []byte) []byte {
	now := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	port := uint32(config.NxosGrpcPort)
	telemetryLogMsg := &telemetry.TelemetryMessageLog{
		SystemId:      &config.ID,
		Location:      &config.Location,
		SourceAddress: &sourceAddress,
		SourcePort:    &port,
		Message: []*telemetry.TelemetryMessage{
			{
				Timestamp: &now,
				Bytes:     data,
			},
		},
	}
	msg, err := proto.Marshal(telemetryLogMsg)
	if err != nil {
		log.Printf("Error cannot serialize telemetry message: %v\n", err)
		return nil
	}
	return msg
}
