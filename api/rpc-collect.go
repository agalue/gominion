package api

import (
	"encoding/xml"
	"strings"
)

// CollectionAttributeDTO represents a collection attribute
type CollectionAttributeDTO struct {
	Key     string `xml:"key,attr"`
	Content string `xml:",innerxml"`
}

// CollectionAgentDTO represents a collection agent
type CollectionAgentDTO struct {
	IPAddress           string                   `xml:"address,attr"`
	StoreByFS           bool                     `xml:"store-by-fs,attr"`
	NodeID              int                      `xml:"node-id,attr"`
	NodeLabel           string                   `xml:"node-label,attr"`
	ForeignSource       string                   `xml:"foreign-source,attr"`
	ForeignID           string                   `xml:"foreign-id,attr"`
	Location            string                   `xml:"location,attr"`
	StorageResourcePath string                   `xml:"storage-resource-path,attr"`
	SysUpTime           int                      `xml:"sys-up-time,attr"`
	Attributes          []CollectionAttributeDTO `xml:"attribute,omitempty"`
}

// CollectorRequestDTO represents a collector request
type CollectorRequestDTO struct {
	XMLName                    xml.Name                 `xml:"collector-request"`
	Location                   string                   `xml:"location,attr"`
	SystemID                   string                   `xml:"system-id,attr"`
	ClassName                  string                   `xml:"class-name,attr"`
	AttributesNeedUnmarshaling bool                     `xml:"attributes-need-unmarshaling,attr"`
	CollectionAgent            *CollectionAgentDTO      `xml:"agent,omitempty"`
	Attributes                 []CollectionAttributeDTO `xml:"attribute,omitempty"`
}

// GetCollector returns the simple class name for the collector implementation
func (req *CollectorRequestDTO) GetCollector() string {
	if req.ClassName == "" {
		return ""
	}
	sections := strings.Split(req.ClassName, ".")
	return sections[len(sections)-1]
}

// CollectorResponseDTO represents a collector response
type CollectorResponseDTO struct {
	XMLName xml.Name `xml:"collector-response"`
	Error   string   `xml:"error,attr"`
}
