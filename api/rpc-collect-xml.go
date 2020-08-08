package api

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
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

// XMLSource represents an XML source
type XMLSource struct {
	XMLName xml.Name    `xml:"xml-source"`
	URL     string      `xml:"url,attr"`
	Groups  []XMLGroup  `xml:"xml-group"`
	Request *XMLRequest `xml:"request,omitempty"`
}

// GetHTTPRequest gets an HTTP request
func (src *XMLSource) GetHTTPRequest() (*http.Request, error) {
	req, err := http.NewRequest(src.getRequest().GetMethod(), src.URL, src.getRequest().GetBody())
	if err != nil {
		return nil, err
	}
	for _, header := range src.getRequest().Headers {
		req.Header.Add(header.Name, header.Value)
	}
	return req, err
}

func (src *XMLSource) getRequest() *XMLRequest {
	if src.Request == nil {
		return &XMLRequest{}
	}
	return src.Request
}

// XMLCollection represents an XML collection object
type XMLCollection struct {
	XMLName xml.Name    `xml:"xml-collection"`
	Name    string      `xml:"name,attr"`
	RRD     *RRD        `xml:"rrd"`
	Sources []XMLSource `xml:"xml-source"`
}
