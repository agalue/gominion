package collectors

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

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
