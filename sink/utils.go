package sink

import (
	"encoding/xml"
	"fmt"
	"net"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/agalue/gominion/protobuf/telemetry"
	"github.com/google/uuid"

	"google.golang.org/protobuf/proto"
)

func sendXMLResponse(moduleID string, config *api.MinionConfig, sink api.Sink, object interface{}) {
	bytes, err := xml.MarshalIndent(object, "", "   ")
	if err != nil {
		log.Errorf("Cannot parse Sink API response: %v", err)
	}
	sendBytes(moduleID, config, sink, bytes)
}

func sendBytes(moduleID string, config *api.MinionConfig, sink api.Sink, bytes []byte) {
	msg := &ipc.SinkMessage{
		MessageId: uuid.New().String(),
		ModuleId:  moduleID,
		SystemId:  config.ID,
		Location:  config.Location,
		Content:   bytes,
	}
	if sink == nil {
		return
	}
	if err := sink.Send(msg); err != nil {
		log.Errorf("%s cannot send message via Sink API: %v", moduleID, err)
	}
}

func wrapMessageToTelemetry(config *api.MinionConfig, sourceAddress string, sourcePort uint32, data [][]byte) []byte {
	now := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	logMsg := &telemetry.TelemetryMessageLog{
		SystemId:      &config.ID,
		Location:      &config.Location,
		SourceAddress: &sourceAddress,
		SourcePort:    &sourcePort,
		Message:       make([]*telemetry.TelemetryMessage, len(data)),
	}
	for i := 0; i < len(data); i++ {
		logMsg.Message[i] = &telemetry.TelemetryMessage{
			Timestamp: &now,
			Bytes:     data[i],
		}
	}
	bytes, err := proto.Marshal(logMsg)
	if err != nil {
		log.Errorf("Cannot serialize telemetry message: %v", err)
		return nil
	}
	return bytes
}

func createUDPListener(port int) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("cannot resolve address: %s", err)
	}
	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on UDP port %d: %s", port, err)
	}
	return conn, nil
}
