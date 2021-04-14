package sink

import (
	"encoding/xml"
	"testing"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"
	"github.com/agalue/gominion/protobuf/telemetry"
	"github.com/golang/protobuf/proto"

	"gotest.tools/v3/assert"
)

type MockSink struct {
	messages []*ipc.SinkMessage
}

func (sink *MockSink) Send(msg *ipc.SinkMessage) error {
	sink.messages = append(sink.messages, msg)
	return nil
}

type Person struct {
	XMLName   xml.Name `xml:"person"`
	FirstName string   `xml:"first-name"`
	LastName  string   `xml:"last-name"`
}

func TestSendXMLResponse(t *testing.T) {
	sink := new(MockSink)
	config := &api.MinionConfig{ID: "minion1", Location: "Test"}
	object := Person{FirstName: "Alejandro", LastName: "Galue"}

	sendXMLResponse("Test", config, sink, object)

	assert.Equal(t, 1, len(sink.messages))
	msg := sink.messages[0]
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

	messages := make([][]byte, 1)
	messages[0] = data
	bytes := wrapMessageToTelemetry(config, ipaddr, port, messages)

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
