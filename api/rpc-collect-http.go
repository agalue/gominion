package api

import (
	"encoding/xml"
)

// HTTPAttribute represents an HTTP attribute
type HTTPAttribute struct {
	XMLName    xml.Name `xml:"attrib"`
	Alias      string   `xml:"alias,attr"`
	MatchGroup int      `xml:"match-group,attr"`
	Type       string   `xml:"type,attr"`
}

// HTTPAttributeList represents a list of HTTP attributes object
type HTTPAttributeList struct {
	XMLName       xml.Name        `xml:"attributes"`
	AttributeList []HTTPAttribute `xml:"attrib"`
}

// HTTPUrl represents an HTTP URL object
type HTTPUrl struct {
	XMLName       xml.Name `xml:"url"`
	Path          string   `xml:"path,attr"`
	UserAgent     string   `xml:"user-agent,attr"`
	Matches       string   `xml:"matches,attr"`
	ResponseRange string   `xml:"response-range,attr"`
}

// HTTPUri represents an HTTP URI object
type HTTPUri struct {
	XMLName    xml.Name           `xml:"uri"`
	Name       string             `xml:"name,attr"`
	URL        *HTTPUrl           `xml:"url"`
	Attributes *HTTPAttributeList `xml:"attributes"`
}

// AddAttribute adds a new HTTP attribute
func (uri *HTTPUri) AddAttribute(attrib HTTPAttribute) {
	if uri.Attributes == nil {
		uri.Attributes = &HTTPAttributeList{}
	}
	uri.Attributes.AttributeList = append(uri.Attributes.AttributeList, attrib)
}

// FindAttributeByMatchGroup finds an attribute by metch group
func (uri *HTTPUri) FindAttributeByMatchGroup(matchGroup int) *HTTPAttribute {
	if uri.Attributes == nil {
		uri.Attributes = &HTTPAttributeList{}
	}
	for _, attr := range uri.Attributes.AttributeList {
		if attr.MatchGroup == matchGroup {
			return &attr
		}
	}
	return nil
}

// HTTPUriList represents a list of HTTP URI object
type HTTPUriList struct {
	XMLName xml.Name  `xml:"uris"`
	URIList []HTTPUri `xml:"uri"`
}

// HTTPCollection represents an HTTP collection object
type HTTPCollection struct {
	XMLName xml.Name     `xml:"http-collection"`
	Name    string       `xml:"name,attr"`
	RRD     *RRD         `xml:"rrd"`
	URIs    *HTTPUriList `xml:"uris"`
}

// AddURI adds a new URI to the HTTP collection
func (collection *HTTPCollection) AddURI(uri HTTPUri) {
	if collection.URIs == nil {
		collection.URIs = &HTTPUriList{}
	}
	collection.URIs.URIList = append(collection.URIs.URIList, uri)
}
