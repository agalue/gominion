package api

import (
	"encoding/xml"
	"testing"

	"gotest.tools/assert"
)

func TestXMLRequest(t *testing.T) {
	xmlstr := `
	<collector-request location="Demo" class-name="org.opennms.protocols.xml.collector.XmlCollector" attributes-need-unmarshaling="true">
		<agent address="192.168.0.18" store-by-fs="true" node-id="2" node-label="agalue-srv01" foreign-source="Demo" foreign-id="agalue-srv01" location="Demo" storage-resource-path="fs/Demo/agalue-srv01" sys-up-time="-1"/>
		<attribute key="collection"><![CDATA[sensors]]></attribute>
		<attribute key="xmlDatacollection"><![CDATA[<xml-collection xmlns="http://xmlns.opennms.org/xsd/config/xml-datacollection" name="sensors">
			<rrd step="300">
			<rra>RRA:AVERAGE:0.5:1:2016</rra>
			<rra>RRA:AVERAGE:0.5:12:1488</rra>
			<rra>RRA:AVERAGE:0.5:288:366</rra>
			<rra>RRA:MAX:0.5:288:366</rra>
			<rra>RRA:MIN:0.5:288:366</rra>
			</rrd>
			<xml-source url="http://192.168.0.18/sensors">
				<xml-group name="sensors" resource-type="node" resource-xpath="/html/body">
					<xml-object name="temperature" type="gauge" xpath="p[@id='temp']"/>
					<xml-object name="humidity" type="gauge" xpath="p[@id='humid']"/>
				</xml-group>
				<import-groups>xml-datacollection/http-sensors.xml</import-groups>
			</xml-source>
		</xml-collection>]]></attribute>
	 </collector-request>
	`

	// Parse and validate HTTP request
	request := &CollectorRequestDTO{}
	err := xml.Unmarshal([]byte(xmlstr), request)
	assert.NilError(t, err)
	assert.Equal(t, "Demo", request.Location)
	assert.Equal(t, "192.168.0.18", request.CollectionAgent.IPAddress)
	assert.Equal(t, "sensors", request.GetAttributeValue("collection", ""))

	// Parse and validate HTTP collection
	collectionParam := request.GetAttributeValue("xmlDatacollection", "")
	collection := &XMLCollection{}
	err = xml.Unmarshal([]byte(collectionParam), collection)
	assert.NilError(t, err)
	assert.Equal(t, "sensors", collection.Name)
	assert.Equal(t, 300, collection.RRD.Step)
	assert.Equal(t, "RRA:AVERAGE:0.5:1:2016", collection.RRD.RRAs[0].Content)
	assert.Equal(t, 1, len(collection.Sources))
	assert.Equal(t, 1, len(collection.Sources[0].Groups))
	assert.Equal(t, 2, len(collection.Sources[0].Groups[0].Objects))
	assert.Equal(t, "temperature", collection.Sources[0].Groups[0].Objects[0].Name)
}
