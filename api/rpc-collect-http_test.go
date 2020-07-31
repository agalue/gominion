package api

import (
	"encoding/xml"
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestHttpConfig(t *testing.T) {
	xmlstr := `
	<http-collection name="doc-count">
		<rrd step="300">
			<rra>RRA:AVERAGE:0.5:1:2016</rra>
			<rra>RRA:AVERAGE:0.5:12:1488</rra>
			<rra>RRA:AVERAGE:0.5:288:366</rra>
			<rra>RRA:MAX:0.5:288:366</rra>
			<rra>RRA:MIN:0.5:288:366</rra>
		</rrd>
		<uris>
			<uri name="document-counts">
				<url path="/httpcolltest.html" user-agent="Mozilla/5.0" matches=".*([0-9]+).*" response-range="100-399"/>
				<attributes>
					<attrib alias="documentCount" match-group="1" type="counter32"/>
				</attributes>
			</uri>
		</uris>
	</http-collection>
	`
	collection := &HTTPCollection{}
	err := xml.Unmarshal([]byte(xmlstr), collection)
	assert.NilError(t, err)
	assert.Equal(t, "doc-count", collection.Name)
	assert.Equal(t, 300, collection.RRD.Step)
	assert.Equal(t, "RRA:AVERAGE:0.5:1:2016", collection.RRD.RRAs[0].Content)
	assert.Equal(t, "document-counts", collection.URIs.URIList[0].Name)
	assert.Equal(t, "100-399", collection.URIs.URIList[0].URL.ResponseRange)
	assert.Equal(t, "documentCount", collection.URIs.URIList[0].Attributes.AttributeList[0].Alias)
}

func TestHTTPRequest(t *testing.T) {
	xmlstr := `
	<collector-request location="Demo" class-name="org.opennms.netmgt.collectd.HttpCollector" attributes-need-unmarshaling="true">
		<agent address="192.168.0.18" store-by-fs="true" node-id="2" node-label="agalue-srv01" foreign-source="Demo" foreign-id="agalue-srv01" location="Demo" storage-resource-path="fs/Demo/agalue-srv01" sys-up-time="-1"/>
		<attribute key="collection"><![CDATA[stats]]></attribute>
		<attribute key="httpCollection"><![CDATA[<http-collection xmlns="http://xmlns.opennms.org/xsd/config/http-datacollection" name="stats">
			<rrd step="30">
			<rra>RRA:AVERAGE:0.5:1:2016</rra>
			<rra>RRA:AVERAGE:0.5:12:1488</rra>
			<rra>RRA:AVERAGE:0.5:288:366</rra>
			<rra>RRA:MAX:0.5:288:366</rra>
			<rra>RRA:MIN:0.5:288:366</rra>
			</rrd>
			<uris>
			<uri name="stats">
				<url user-agent="Mozilla/5.0 (Macintosh; U; PPC Mac OS X; en) AppleWebKit/412 (KHTML, like Gecko) Safari/412" path="/stats" matches="(?s).*Temperature: ([0-9]+).*Humidity: ([0-9]+)" response-range="100-399"/>
				<attributes>
					<attrib alias="temperature" match-group="1" type="gauge"/>
					<attrib alias="humidity" match-group="2" type="gauge"/>
				</attributes>
			</uri>
			</uris>
		</http-collection>]]></attribute>
	</collector-request>
	`
	request := &CollectorRequestDTO{}
	err := xml.Unmarshal([]byte(xmlstr), request)
	assert.NilError(t, err)
	collectionParam := request.GetAttributeValue("httpCollection")
	collection := &HTTPCollection{}
	err = xml.Unmarshal([]byte(collectionParam), collection)
	assert.NilError(t, err)
	assert.Equal(t, "stats", collection.Name)
	fmt.Println(collectionParam)
}
