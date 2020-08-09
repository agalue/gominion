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
		<uri name="sensors">
			<url path="/stats"
				user-agent="Mozilla/5.0 (Macintosh; U; PPC Mac OS X; en)"
				matches="(?s).*Temperature: ([0-9]+).*Humidity: ([0-9]+).*"
				response-range="100-399"/>
			<attributes>
				<attrib alias="temperature" match-group="1" type="gauge32"/>
				<attrib alias="humidity"    match-group="2" type="gauge32"/>
			</attributes>
		</uri>
	</uris>
</http-collection>
`

var mockHTML = `
<html>
	<p>Temperature: 29</p>
	<p>Humidity: 66</p>
</html>
`

func TestAddResourceAttributes(t *testing.T) {
	collection := &api.HTTPCollection{}
	err := xml.Unmarshal([]byte(mockHTTPCollection), collection)
	assert.NilError(t, err)
	uri := collection.URIs.URIList[0]
	resource := &api.CollectionResourceDTO{Name: "node"}
	builder := api.NewCollectionSetBuilder(nil)
	err = httpCollector.AddResourceAttributes(builder, resource, uri, mockHTML)
	assert.NilError(t, err)
	cs := builder.Build()
	assert.Equal(t, 1, len(cs.Resources))
	assert.Equal(t, 2, len(cs.Resources[0].NumericAttributes))
	assert.Equal(t, "29", cs.Resources[0].NumericAttributes[0].Value)
	assert.Equal(t, "66", cs.Resources[0].NumericAttributes[1].Value)
}

func TestHttpCollector(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(mockHTML))
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
			{Key: httpCollectionAttr, Content: mockHTTPCollection},
			{Key: "port", Content: u.Port()},
		},
	}
	response := httpCollector.Collect(request)
	bytes, err := xml.MarshalIndent(response, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
	assert.Equal(t, api.CollectionStatusSucceded, response.CollectionSet.Status)
	assert.Equal(t, 1, len(response.CollectionSet.Resources))
	assert.Equal(t, 2, len(response.CollectionSet.Resources[0].NumericAttributes))
}
