package collectors

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/tools"
)

// HTTPCollector represents a collector implementation
type HTTPCollector struct {
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (collector *HTTPCollector) GetID() string {
	return "HttpCollector"
}

// Collect execute the collector request and return the collection set
func (collector *HTTPCollector) Collect(request *api.CollectorRequestDTO) api.CollectorResponseDTO {
	response := api.CollectorResponseDTO{
		CollectionSet: &api.CollectionSetDTO{
			Timestamp: &api.Timestamp{Time: time.Now()},
			Status:    api.CollectionStatusFailed,
			Agent:     request.CollectionAgent,
		},
	}

	httpCollection := &api.HTTPCollection{}
	err := xml.Unmarshal([]byte(request.GetAttributeValue("httpCollection", "")), httpCollection)
	if err != nil {
		response.Error = fmt.Sprintf("Error cannot parse httpCollection: %s", err.Error())
		return response
	}

	nodeResource := api.CollectionResourceDTO{Name: "node"}
	for _, uri := range httpCollection.URIs.URIList {
		u := url.URL{
			Scheme: "http",
			Host:   request.CollectionAgent.IPAddress + ":" + request.GetAttributeValue("port", "80"),
			Path:   uri.URL.Path,
		}
		log.Printf("Executing an HTTP GET against %s", u.String())
		httpreq, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			response.Error = err.Error()
			return response
		}
		if uri.URL.UserAgent != "" {
			httpreq.Header.Set("User-Agent", uri.URL.UserAgent)
		}
		var timeout time.Duration = 3 * time.Second
		if t := request.GetAttributeValue("timeout", "3000"); t != "" {
			if v, err := strconv.Atoi(t); err != nil {
				timeout = time.Duration(v) * time.Millisecond
			}
		}
		client := tools.GetHTTPClient(false, timeout)
		httpres, err := client.Do(httpreq)
		if err != nil {
			response.Error = err.Error()
			return response
		}
		min, max := tools.ParseHTTPResponseRange(uri.URL.ResponseRange)
		if httpres.StatusCode < min || httpres.StatusCode > max {
			response.Error = fmt.Sprintf("Response code %d out of expected range: %d-%d", httpres.StatusCode, min, max)
			return response
		}
		data, err := ioutil.ReadAll(httpres.Body)
		if err != nil {
			response.Error = err.Error()
			return response
		}
		collector.AddResourceAttributes(&nodeResource, uri, string(data))
	}

	response.CollectionSet.Status = api.CollectionStatusSucceded
	response.CollectionSet.AddResource(nodeResource)
	return response
}

// AddResourceAttributes adds attributes to resource based on HTML and URI configuration
func (collector *HTTPCollector) AddResourceAttributes(resource *api.CollectionResourceDTO, uri api.HTTPUri, html string) error {
	rp, err := regexp.Compile(uri.URL.Matches)
	if err != nil {
		return err
	}
	groups := rp.FindAllStringSubmatch(html, -1)
	if len(groups) == 1 {
		for i := 1; i < len(groups[0]); i++ {
			if attr := uri.FindAttributeByMatchGroup(i); attr != nil {
				resource.AddAttribute(attr.Type, attr.Alias, groups[0][i])
			}
		}
	}
	return nil
}

var httpCollector = &HTTPCollector{}

func init() {
	RegisterCollector(httpCollector)
}
