package api

import (
	"encoding/xml"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHTTPRequest(t *testing.T) {
	xmlstr := `
	<collector-request location="Demo" class-name="org.opennms.netmgt.collectd.HttpCollector" attributes-need-unmarshaling="true">
		<agent address="192.168.0.18" store-by-fs="true" node-id="2" node-label="agalue-srv01" foreign-source="Demo" foreign-id="agalue-srv01" location="Demo" storage-resource-path="fs/Demo/agalue-srv01" sys-up-time="-1"/>
		<attribute key="collection"><![CDATA[stats]]></attribute>
		<attribute key="httpCollection"><![CDATA[<http-collection xmlns="http://xmlns.opennms.org/xsd/config/http-datacollection" name="stats">
			<rrd step="300">
				<rra>RRA:AVERAGE:0.5:1:2016</rra>
				<rra>RRA:AVERAGE:0.5:12:1488</rra>
				<rra>RRA:AVERAGE:0.5:288:366</rra>
				<rra>RRA:MAX:0.5:288:366</rra>
				<rra>RRA:MIN:0.5:288:366</rra>
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
		</http-collection>]]></attribute>
	</collector-request>
	`

	// Parse and validate HTTP request
	request := &CollectorRequestDTO{}
	err := xml.Unmarshal([]byte(xmlstr), request)
	assert.NilError(t, err)
	assert.Equal(t, "Demo", request.Location)
	assert.Equal(t, "192.168.0.18", request.CollectionAgent.IPAddress)
	assert.Equal(t, "stats", request.GetAttributeValue("collection", ""))

	// Parse and validate HTTP collection
	collectionParam := request.GetAttributeValue("httpCollection", "")
	collection := &HTTPCollection{}
	err = xml.Unmarshal([]byte(collectionParam), collection)
	assert.NilError(t, err)
	assert.Equal(t, "stats", collection.Name)
	assert.Equal(t, 300, collection.RRD.Step)
	assert.Equal(t, "RRA:AVERAGE:0.5:1:2016", collection.RRD.RRAs[0].Content)
	uri := collection.FindURI("sensors")
	assert.Assert(t, uri != nil)
	assert.Equal(t, "100-399", uri.URL.ResponseRange)
	assert.Equal(t, 2, len(uri.Attributes.AttributeList))
	temperature := uri.FindAttributeByAlias("temperature")
	assert.Assert(t, temperature != nil)
	assert.Equal(t, "temperature", temperature.Alias)
	humidity := uri.FindAttributeByMatchGroup(2)
	assert.Assert(t, humidity != nil)
	assert.Equal(t, "humidity", humidity.Alias)
}
