package sink

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"testing"
	"time"

	"github.com/agalue/gominion/api"

	"gotest.tools/v3/assert"
)

func TestSyslogBuildMessageLog(t *testing.T) {
	logParts := make(map[string]interface{})
	logParts["client"] = "127.0.0.1:64557"
	logParts["content"] = ": 2020 Jul 29 17:07:29 EDT: %ETHPORT-5-IF_DOWN_LINK_FAILURE: Interface eth1 is down (Link failure)"
	logParts["facility"] = 23
	logParts["hostname"] = "127.0.0.1"
	logParts["priority"] = 189
	logParts["severity"] = 5
	logParts["tags"] = ""
	logParts["timestamp"] = time.Now()
	logParts["tls_peer"] = ""

	module := &SyslogModule{
		config: &api.MinionConfig{
			ID:       "minion1",
			Location: "Test",
		},
	}

	logMsg := module.buildMessageLog(logParts)
	bytes, err := xml.Marshal(logMsg)
	assert.NilError(t, err)
	fmt.Println(string(bytes))

	if logMsg == nil {
		t.FailNow()
	} else {
		assert.Equal(t, "Test", logMsg.Location)
		assert.Equal(t, 1, len(logMsg.Messages))
		decodedMsg, err := base64.StdEncoding.DecodeString(string(logMsg.Messages[0].Content))
		assert.NilError(t, err)
		expectedMsg := fmt.Sprintf("<%d>%s", logParts["priority"], logParts["content"])
		fmt.Printf("Expected message: %s\n", expectedMsg)
		assert.Equal(t, expectedMsg, string(decodedMsg))
	}

	logParts["content"] = "X"
	logMsg = module.buildMessageLog(logParts)
	assert.Assert(t, logMsg == nil)
}
