package collectors

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/agalue/gominion/api"
	"gotest.tools/assert"
)

var mockHTTPCollection = `
<http-collection name="sample">
	<rrd step="300">
		<rra>RRA:AVERAGE:0.5:1:2016</rra>
	</rrd>
	<uris>
		<uri name="sample">
			<url path="/" user-agent="Mozilla/5.0" matches="(?s).*var1=([0-9]+).*var2=([0-9]+).*" response-range="100-399"/>
			<attributes>
				<attrib alias="var1" match-group="1" type="counter"/>
				<attrib alias="var2" match-group="2" type="gauge"/>
			</attributes>
		</uri>
	</uris>
</http-collection>
`

var mockHTML = `
<html>
	<p>var1=10</p>
	<p>var2=20</p>
</html>
`

func TestAddResourceAttributes(t *testing.T) {
	collection := &api.HTTPCollection{}
	err := xml.Unmarshal([]byte(mockHTTPCollection), collection)
	assert.NilError(t, err)
	uri := collection.URIs.URIList[0]
	resource := &api.CollectionResourceDTO{Name: "node"}
	err = httpCollector.AddResourceAttributes(resource, uri, mockHTML)
	assert.NilError(t, err)
	assert.Equal(t, 2, len(resource.Attributes))
	assert.Equal(t, "10", resource.Attributes[0].Value.Content)
	assert.Equal(t, "20", resource.Attributes[1].Value.Content)
}

func TestHttpCollector(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(mockHTML))
		res.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	u, err := url.Parse(testServer.URL)
	assert.NilError(t, err)
	fmt.Printf("Mock Server Hostname: %s, Port: %s\n", u.Hostname(), u.Port())
	request := &api.CollectorRequestDTO{
		CollectionAgent: &api.CollectionAgentDTO{
			IPAddress:           u.Hostname(),
			NodeID:              1,
			NodeLabel:           "srv01",
			StorageResourcePath: "snmp/1/node",
			SysUpTime:           time.Now().Unix(),
		},
		Attributes: []api.CollectionAttributeDTO{
			{Key: "httpCollection", Content: mockHTTPCollection},
			{Key: "port", Content: u.Port()},
		},
	}
	response := httpCollector.Collect(request)
	bytes, err := xml.MarshalIndent(response, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
	assert.Equal(t, api.CollectionStatusSucceded, response.CollectionSet.Status)
}
