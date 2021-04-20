package collectors

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/tools"
)

const httpCollectionAttr = "httpCollection"

// HTTPCollector represents a collector implementation
type HTTPCollector struct {
}

// GetID gets the collector ID (simple class name from its Java counterpart)
func (collector *HTTPCollector) GetID() string {
	return "HttpCollector"
}

// Collect execute the collector request and return the collection response
func (collector *HTTPCollector) Collect(request *api.CollectorRequestDTO) *api.CollectorResponseDTO {
	response := &api.CollectorResponseDTO{}
	httpCollection := &api.HTTPCollection{}
	err := xml.Unmarshal([]byte(request.GetAttributeValue(httpCollectionAttr, "")), httpCollection)
	if err != nil {
		response.MarkAsFailed(request.CollectionAgent, fmt.Errorf("cannot parse %s: %v", httpCollectionAttr, err))
		return response
	}
	builder := api.NewCollectionSetBuilder(request.CollectionAgent)
	nodeResource := &api.CollectionResourceDTO{
		ResourceType: &api.NodeLevelResourceDTO{
			NodeID: request.CollectionAgent.NodeID,
		},
	}
	for _, uri := range httpCollection.URIs.URIList {
		u := url.URL{
			Scheme: "http",
			Host:   request.CollectionAgent.IPAddress + ":" + request.GetAttributeValue("port", "80"),
			Path:   uri.URL.Path,
		}
		log.Debugf("Executing an HTTP GET against %s", u.String())
		httpreq, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			response.MarkAsFailed(request.CollectionAgent, err)
			return response
		}
		if uri.URL.UserAgent != "" {
			httpreq.Header.Set("User-Agent", uri.URL.UserAgent)
		}
		client := tools.GetHTTPClient(false, request.GetTimeout())
		httpres, err := client.Do(httpreq)
		if err != nil {
			response.MarkAsFailed(request.CollectionAgent, err)
			return response
		}
		min, max := tools.ParseHTTPResponseRange(uri.URL.ResponseRange)
		if httpres.StatusCode < min || httpres.StatusCode > max {
			exerr := fmt.Errorf("response code %d out of expected range: %d-%d", httpres.StatusCode, min, max)
			response.MarkAsFailed(request.CollectionAgent, exerr)
			return response
		}
		data, err := ioutil.ReadAll(httpres.Body)
		if err != nil {
			response.MarkAsFailed(request.CollectionAgent, err)
			return response
		}
		collector.AddResourceAttributes(builder, nodeResource, uri, string(data))
	}
	response.CollectionSet = builder.Build()
	return response
}

// AddResourceAttributes adds attributes to resource based on HTML and URI configuration
func (collector *HTTPCollector) AddResourceAttributes(builder *api.CollectionSetBuilder, cres *api.CollectionResourceDTO, uri api.HTTPUri, html string) error {
	rp, err := regexp.Compile(uri.URL.Matches)
	if err != nil {
		return err
	}
	groups := rp.FindAllStringSubmatch(html, -1)
	if len(groups) == 1 {
		for i := 1; i < len(groups[0]); i++ {
			if attr := uri.FindAttributeByMatchGroup(i); attr != nil {
				builder.WithAttribute(cres, uri.Name, attr.Alias, groups[0][i], attr.Type)
			}
		}
	}
	return nil
}

func init() {
	RegisterCollector(&HTTPCollector{})
}
