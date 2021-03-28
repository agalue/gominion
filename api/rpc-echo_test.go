package api

import (
	"encoding/xml"
	"testing"

	"gotest.tools/v3/assert"
)

func TestEchoRequest(t *testing.T) {
	requestXML := `
	<echo-request id="10" message="Test Message" location="Test" system-id="minion1" delay="100" throw="false">
		<body>Test Body</body>
	</echo-request>
	`

	request := &EchoRequest{}
	err := xml.Unmarshal([]byte(requestXML), request)
	assert.NilError(t, err)
	assert.Equal(t, "Test", request.Location)
	assert.Equal(t, "Test Message", request.Message)
	assert.Equal(t, "Test Body", request.Body)
}

func TestEchoResponse(t *testing.T) {
	responseXML := `
	<echo-response id="10" message="Test Message">
		<body>Test Body</body>
	</echo-response>
	`

	response := &EchoResponse{}
	err := xml.Unmarshal([]byte(responseXML), response)
	assert.NilError(t, err)
	assert.Equal(t, "Test Message", response.Message)
	assert.Equal(t, "Test Body", response.Body)
}

func TestEchoErrorResponse(t *testing.T) {
	responseXML := `<echo-response id="10" error="Something went wrong"></echo-response>`

	response := &EchoResponse{}
	err := xml.Unmarshal([]byte(responseXML), response)
	assert.NilError(t, err)
	assert.Equal(t, "Something went wrong", response.Error)
}
