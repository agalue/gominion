package api

import (
	"encoding/xml"
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestEchoRequest(t *testing.T) {
	request := &EchoRequest{
		ID:       10,
		Delay:    100,
		Location: "Test",
		SystemID: "minion1",
		Message:  "Test Message",
		Body:     "Test Body",
	}
	bytes, err := xml.MarshalIndent(request, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
}

func TestEchoResponse(t *testing.T) {
	response := &EchoResponse{
		ID:      10,
		Message: "Test Message",
		Body:    "Test Body",
	}
	bytes, err := xml.MarshalIndent(response, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))

	response = &EchoResponse{
		ID:    10,
		Error: "Something went wrong",
	}
	bytes, err = xml.MarshalIndent(response, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
}
