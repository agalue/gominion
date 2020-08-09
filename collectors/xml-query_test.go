package collectors

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/agalue/gominion/api"
	"gotest.tools/assert"
)

var queryTestXML = `
<zones>
	<zone id="0" name="global" timestamp="1299258742">
		<parameter key="nproc" value="245" />
		<parameter key="nlwp" value="1455" />
		<parameter key="pr_size" value="2646864" />
		<parameter key="pr_rssize" value="1851072" />
		<parameter key="pctmem" value="0.7" />
		<parameter key="pctcpu" value="0.24" />
	</zone>
	<zone id="871" name="zone1" timestamp="1299258742">
		<parameter key="nproc" value="24" />
		<parameter key="nlwp" value="328" />
		<parameter key="pr_size" value="1671128" />
		<parameter key="pr_rssize" value="1193240" />
		<parameter key="pctmem" value="0.4" />
		<parameter key="pctcpu" value="0.07" />
	</zone>
</zones>`

var queryTestJSON = `
{
	"zones": [
		{
			"id" : 0,
			"name" : "global",
			"timestamp" : 1299258742,
			"parameter" : [
				{ "key" : "nproc", "value" : "245" },
				{ "key" : "nlwp", "value" : "1455" },
				{ "key" : "pr_size", "value" : "2646864" },
				{ "key" : "pr_rssize", "value" : "1851072" },
				{ "key" : "pctmem", "value" : "0.7" },
				{ "key" : "pctcpu", "value" : "0.24" }
			]
		}, {
			"id" : 871,
			"name" : "zone1",
			"timestamp" : 1299258742,
			"parameter" : [
				{ "key" : "nproc", "value" : "24" },
				{ "key" : "nlwp", "value" : "328" },
				{ "key" : "pr_size", "value" : "1671128" },
				{ "key" : "pr_rssize", "value" : "1193240" },
				{ "key" : "pctmem", "value" : "0.4" },
				{ "key" : "pctcpu", "value" : "0.07" }
			]
		}
	]
}`

var queryTestHTML = `
<html>
	<head>
		<title>Sensor Stats</title>
	</head>
	<body>
		<img src="/logo.jpg">
		<hr>
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

func TestXMLQuerier(t *testing.T) {
	querier, err := NewQuerier(XMLHandlerClass, nil)
	assert.NilError(t, err)
	doc, err := querier.Parse(strings.NewReader(queryTestXML))
	assert.NilError(t, err)
	zones, err := querier.QueryAll(doc, "/zones/*")
	assert.NilError(t, err)
	assert.Equal(t, 2, len(zones))
	for _, zone := range zones {
		name, err := querier.Query(zone, "@name")
		assert.NilError(t, err)
		nlwp, err := querier.Query(zone, "parameter[@key='nlwp']/@value")
		assert.NilError(t, err)
		re, err := regexp.Compile("[.\\d]+")
		assert.NilError(t, err)
		data := re.FindAllString(nlwp.GetContent(), -1)
		assert.Equal(t, 1, len(data))
		fmt.Printf("%s.blwp=%s\n", name.GetContent(), data[0])
	}
}

func TestJSONQuerier(t *testing.T) {
	querier, err := NewQuerier(JSONHandlerClass, nil)
	assert.NilError(t, err)
	doc, err := querier.Parse(strings.NewReader(queryTestJSON))
	assert.NilError(t, err)
	zones, err := querier.QueryAll(doc, "/zones/*")
	assert.NilError(t, err)
	assert.Equal(t, 2, len(zones))
	for _, zone := range zones {
		name, err := querier.Query(zone, "name")
		assert.NilError(t, err)
		nlwp, err := querier.Query(zone, "parameter/*[key='nlwp']/value")
		assert.NilError(t, err)
		re, err := regexp.Compile("[.\\d]+")
		assert.NilError(t, err)
		data := re.FindAllString(nlwp.GetContent(), -1)
		assert.Equal(t, 1, len(data))
		fmt.Printf("%s.blwp=%s\n", name.GetContent(), data[0])
	}
}

func TestHTMLQuerier(t *testing.T) {
	querier, err := NewQuerier("", &api.XMLRequest{
		Parameters: []api.XMLRequestParameter{
			{Name: "pre-parse-html", Value: "true"},
		}})
	assert.NilError(t, err)
	doc, err := querier.Parse(strings.NewReader(queryTestHTML))
	assert.NilError(t, err)
	body, err := querier.Query(doc, "/html/body")
	assert.NilError(t, err)
	temp, err := querier.Query(body, "p[@id='temp']")
	assert.NilError(t, err)
	re, err := regexp.Compile("[.\\d]+")
	assert.NilError(t, err)
	data := re.FindAllString(temp.GetContent(), -1)
	assert.Equal(t, 1, len(data))
	assert.Equal(t, "92.5", data[0])
	resources, err := querier.QueryAll(body, "//tr")
	assert.NilError(t, err)
	assert.Equal(t, 3, len(resources))
	for _, resource := range resources {
		name, err := querier.Query(resource, "td[@id='rack']")
		assert.NilError(t, err)
		value, err := querier.Query(resource, "td[@id='fan']")
		fmt.Printf("%s = %s\n", name.GetContent(), value.GetContent())
	}
}

func TestCSSQuerier(t *testing.T) {
	querier, err := NewQuerier(HTTPHandlerClass, nil)
	assert.NilError(t, err)
	doc, err := querier.Parse(strings.NewReader(queryTestHTML))
	assert.NilError(t, err)
	body, err := querier.Query(doc, "html > body")
	assert.NilError(t, err)
	temp, err := querier.Query(body, "p[id='temp']")
	assert.NilError(t, err)
	re, err := regexp.Compile("[.\\d]+")
	assert.NilError(t, err)
	data := re.FindAllString(temp.GetContent(), -1)
	assert.Equal(t, 1, len(data))
	assert.Equal(t, "92.5", data[0])
	resources, err := querier.QueryAll(body, "tr")
	assert.NilError(t, err)
	assert.Equal(t, 3, len(resources))
	for _, resource := range resources {
		name, err := querier.Query(resource, "td[id='rack']")
		assert.NilError(t, err)
		value, err := querier.Query(resource, "td[id='fan']")
		fmt.Printf("%s = %s\n", name.GetContent(), value.GetContent())
	}
}
