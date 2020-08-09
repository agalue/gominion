package collectors

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/agalue/gominion/api"
	"gotest.tools/assert"
)

var mockXMLCollection = `
<xml-collection xmlns="http://xmlns.opennms.org/xsd/config/xml-datacollection" name="sensors">
	<rrd step="300">
	<rra>RRA:AVERAGE:0.5:1:2016</rra>
	<rra>RRA:AVERAGE:0.5:12:1488</rra>
	<rra>RRA:AVERAGE:0.5:288:366</rra>
	<rra>RRA:MAX:0.5:288:366</rra>
	<rra>RRA:MIN:0.5:288:366</rra>
	</rrd>
	<xml-source url="MOCK_URL_HERE">
		<xml-group name="sensors" resource-type="node" resource-xpath="/html/body">
			<xml-object name="temperature" type="gauge" xpath="p[@id='temp']"/>
			<xml-object name="humidity" type="gauge" xpath="p[@id='humid']"/>
		</xml-group>
		<xml-group name="racks" resource-type="node" resource-type="sensorFan" resource-xpath="/html/body/table/tr" key-xpath="td[@id='rack']">
			<xml-object name="fan" type="gauge" xpath="td[@id='fan']"/>
		</xml-group>
		<import-groups>xml-datacollection/sensors.xml</import-groups>
	</xml-source>
</xml-collection>
`

var mockXML = `
<html>
	<head>
		<title>Sensor Stats</title>
	</head>
	<body>
		<p id="temp">Temperature = 92.5</p>
		<p id="humid">Humidity = 66.6</p>
		<table id="racks">
			<tr><td id="rack">Rack 1</td><td id="fan">5200</td></tr>
			<tr><td id="rack">Rack 2</td><td id="fan">4500</td></tr>
			<tr><td id="rack">Rack 3</td><td id="fan">7200</td></tr>
		</table>
	</body>
</html>
`

func TestXmlCollector(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(mockXML))
	}))
	defer testServer.Close()
	u, err := url.Parse(testServer.URL)
	assert.NilError(t, err)
	fmt.Printf("Mock Server Hostname: %s, Port: %s\n", u.Hostname(), u.Port())
	collection := strings.Replace(mockXMLCollection, "MOCK_URL_HERE", u.String(), -1)
	request := &api.CollectorRequestDTO{
		CollectionAgent: &api.CollectionAgentDTO{
			IPAddress:           u.Hostname(),
			NodeID:              1,
			NodeLabel:           "srv01",
			StorageResourcePath: "snmp/1/node",
			SysUpTime:           time.Now().Unix(),
		},
		Attributes: []api.CollectionAttributeDTO{
			{Key: xmlCollectionAttr, Content: collection},
			{Key: "port", Content: u.Port()},
		},
	}
	response := xmlCollector.Collect(request)
	bytes, err := xml.MarshalIndent(response, "", "  ")
	assert.NilError(t, err)
	fmt.Println(string(bytes))
	assert.Equal(t, api.CollectionStatusSucceded, response.CollectionSet.Status)
	assert.Equal(t, 4, len(response.CollectionSet.Resources))
	for _, r := range response.CollectionSet.Resources {
		switch o := r.ResourceType.(type) {
		case *api.NodeLevelResourceDTO:
			fmt.Println("Found node-level resource")
			assert.Equal(t, 2, len(r.NumericAttributes))
		case *api.GenericTypeResourceDTO:
			fmt.Println("Found generic-index resource")
			assert.Equal(t, o.Name, "sensorFan")
			assert.Equal(t, 1, len(r.NumericAttributes))
		default:
			t.FailNow()
		}
	}
}
