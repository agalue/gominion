package sink

import (
	"encoding/xml"
	"testing"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/protobuf/ipc"

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
