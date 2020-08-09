package api

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// XMLObject represents an XML object
type XMLObject struct {
	XMLName xml.Name `xml:"xml-object"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
	XPath   string   `xml:"xpath,attr"`
}

// XMLResourceKey represents a resource key list inside a group
type XMLResourceKey struct {
	XMLName   xml.Name `xml:"resource-key"`
	KeyXPaths []string `xml:"key-xpath"`
}

// XMLGroup represents an XML group
type XMLGroup struct {
	XMLName         xml.Name        `xml:"xml-group"`
	Name            string          `xml:"name,attr"`
	ResourceType    string          `xml:"resource-type,attr"`
	ResourceXPath   string          `xml:"resource-xpath,attr"`
	KeyXPath        string          `xml:"key-xpath,attr"`
	TimestampFormat string          `xml:"timestamp-format,attr"`
	TimestampXPath  string          `xml:"timestamp-xpath,attr"`
	Objects         []XMLObject     `xml:"xml-object,omitempty"`
	ResourceKey     *XMLResourceKey `xml:"resource-key,omitempty"`
}

// HasMultipleResourceKeys checks if the group has multiple resource keys
func (group *XMLGroup) HasMultipleResourceKeys() bool {
	return group.ResourceXPath != "" && group.ResourceKey != nil && len(group.ResourceKey.KeyXPaths) > 0
}

// XMLRequestParameter represents an XML request parameter
type XMLRequestParameter struct {
	XMLName xml.Name `xml:"parameter"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

// XMLRequestHeader represents an XML request header
type XMLRequestHeader struct {
	XMLName xml.Name `xml:"header"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

// XMLRequestContent represents an XML request content
type XMLRequestContent struct {
	XMLName xml.Name `xml:"content"`
	Type    string   `xml:"content-type,attr"`
	Value   string   `xml:",innerxml"`
}

// XMLRequest represents an XML request
type XMLRequest struct {
	XMLName    xml.Name              `xml:"request"`
	Method     string                `xml:"method,attr"`
	Headers    []XMLRequestHeader    `xml:"header,omitempty"`
	Parameters []XMLRequestParameter `xml:"parameter,omitempty"`
	Content    *XMLRequestContent    `xml:"content,omitempty"`
}

// GetMethod returns the HTTP method of the XML Request
func (r *XMLRequest) GetMethod() string {
	if r.Method == "" {
		return "GET"
	}
	return r.Method
}

// GetBody returns a reader with the request body if exist
func (r *XMLRequest) GetBody() io.Reader {
	if r.Content != nil {
		return bytes.NewBufferString(r.Content.Value)
	}
	return nil
}

// GetParameterAsInt gets the value of a request parameter as an integer
func (r *XMLRequest) GetParameterAsInt(key string) int {
	val := r.GetParameterAsString(key)
	v, err := strconv.Atoi(val)
	if err == nil {
		return v
	}
	return 0
}

// GetParameterAsString gets the value of a request parameter as a string
func (r *XMLRequest) GetParameterAsString(key string) string {
	for _, p := range r.Parameters {
		if p.Name == key {
			return p.Value
		}
	}
	return ""
}

// XMLSource represents an XML source
type XMLSource struct {
	XMLName xml.Name    `xml:"xml-source"`
	URL     string      `xml:"url,attr"`
	Groups  []XMLGroup  `xml:"xml-group"`
	Request *XMLRequest `xml:"request,omitempty"`
}

// GetHTTPRequest gets an HTTP request
func (src *XMLSource) GetHTTPRequest() (*http.Request, error) {
	req := src.GetRequest()
	httpreq, err := http.NewRequest(req.GetMethod(), src.URL, req.GetBody())
	if err != nil {
		return nil, err
	}
	for _, header := range req.Headers {
		httpreq.Header.Add(header.Name, header.Value)
	}
	return httpreq, err
}

// GetRequest gets the HTTP request
func (src *XMLSource) GetRequest() *XMLRequest {
	if src.Request == nil {
		return &XMLRequest{}
	}
	return src.Request
}

// SkipSSL checks whether or not to skip certificate validation for HTTPS
func (src *XMLSource) SkipSSL() bool {
	v := src.GetRequest().GetParameterAsString("disable-ssl-verification")
	return strings.ToLower(v) == "true"
}

// XMLCollection represents an XML collection object
type XMLCollection struct {
	XMLName xml.Name    `xml:"xml-collection"`
	Name    string      `xml:"name,attr"`
	RRD     *RRD        `xml:"rrd"`
	Sources []XMLSource `xml:"xml-source"`
}
