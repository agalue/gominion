package api

import (
	"encoding/xml"
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
