package sink

import (
	"encoding/xml"
	"testing"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/agalue/gominion/protobuf/telemetry"
	"github.com/golang/protobuf/proto"

	"gotest.tools/assert"
)

type MockBroker struct {
	messages []*ipc.SinkMessage
}

func (broker *MockBroker) Send(msg *ipc.SinkMessage) error {
	broker.messages = append(broker.messages, msg)
	return nil
}

type Person struct {
	XMLName   xml.Name `xml:"person"`
	FirstName string   `xml:"first-name"`
	LastName  string   `xml:"last-name"`
}

func TestSendResponse(t *testing.T) {
	broker := new(MockBroker)
	config := &api.MinionConfig{ID: "minion1", Location: "Test"}
	object := Person{FirstName: "Alejandro", LastName: "Galue"}

	sendResponse("Test", config, broker, object)

	assert.Equal(t, 1, len(broker.messages))
	msg := broker.messages[0]
	assert.Equal(t, config.ID, msg.SystemId)

	received := &Person{}
	err := xml.Unmarshal(msg.Content, received)
	assert.NilError(t, err)
	assert.Equal(t, object.FirstName, received.FirstName)
}

func TestWrapMessageToTelemetry(t *testing.T) {
	port := uint32(50000)
	ipaddr := "10.0.0.1"
	config := &api.MinionConfig{ID: "minion1", Location: "Test"}
	object := Person{FirstName: "Alejandro", LastName: "Galue"}
	data, err := xml.Marshal(object)
	assert.NilError(t, err)

	bytes := wrapMessageToTelemetry(config, ipaddr, port, data)

	telemetry := &telemetry.TelemetryMessageLog{}
	err = proto.Unmarshal(bytes, telemetry)
	assert.NilError(t, err)
	assert.Equal(t, port, telemetry.GetSourcePort())
	assert.Equal(t, ipaddr, telemetry.GetSourceAddress())
	assert.Equal(t, 1, len(telemetry.Message))

	msg := telemetry.Message[0]
	received := &Person{}
	err = xml.Unmarshal(msg.Bytes, received)
	assert.NilError(t, err)
	assert.Equal(t, object.FirstName, received.FirstName)
}
