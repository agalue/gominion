package rpc

import (
	"encoding/xml"
	"testing"

	"github.com/agalue/gominion/protobuf/ipc"

	"gotest.tools/v3/assert"
)

type Person struct {
	XMLName   xml.Name `xml:"person"`
	FirstName string   `xml:"first-name"`
	LastName  string   `xml:"last-name"`
}

func TestTransformResponse(t *testing.T) {
	object := Person{FirstName: "Alejandro", LastName: "Galue"}
	request := &ipc.RpcRequestProto{
		ModuleId: "Test",
		RpcId:    "001",
		SystemId: "minion1",
		Location: "Test",
	}
	response := transformResponse(request, object)
	if response == nil {
		t.FailNow()
	} else {
		received := &Person{}
		err := xml.Unmarshal(response.RpcContent, received)
		assert.NilError(t, err)
		assert.Equal(t, object.FirstName, received.FirstName)
	}
}
