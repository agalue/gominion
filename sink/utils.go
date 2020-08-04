package sink

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"
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

func wrapMessageToTelemetry(config *api.MinionConfig, sourceAddress string, sourcePort uint32, data []byte) []byte {
	now := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	logMsg := &telemetry.TelemetryMessageLog{
		SystemId:      &config.ID,
		Location:      &config.Location,
		SourceAddress: &sourceAddress,
		SourcePort:    &sourcePort,
		Message: []*telemetry.TelemetryMessage{
			{
				Timestamp: &now,
				Bytes:     data,
			},
		},
	}
	bytes, err := proto.Marshal(logMsg)
	if err != nil {
		log.Printf("Error cannot serialize telemetry message: %v\n", err)
		return nil
	}
	return bytes
}

func startUDPServer(name string, port int) *net.UDPConn {
	log.Printf("Starting %s UDP Forwarder Module", name)
	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Error cannot resolve address for %s: %s", name, err)
	}
	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Fatalf("Error cannot start %s UDP Forwarder: %s", name, err)
	}
	return conn
}
