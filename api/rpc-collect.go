package api

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// CollectionStatusUnknown status unknown
const CollectionStatusUnknown = "UNKNOWN"

// CollectionStatusSucceded collection finished successfully
const CollectionStatusSucceded = "SUCCEEDED"

// CollectionStatusFailed collection failed
const CollectionStatusFailed = "FAILED"

// CollectionAttributeValueDTO represents a collection resource attribute value
type CollectionAttributeValueDTO struct {
	XMLName xml.Name `xml:"value"`
	Type    string   `xml:"type,attr"`
	Content string   `xml:",chardata"`
}

// CollectionResourceAttributeDTO represents a collection resource attribute
type CollectionResourceAttributeDTO struct {
	XMLName xml.Name                     `xml:"attribute"`
	Name    string                       `xml:"name,attr"`
	Value   *CollectionAttributeValueDTO `xml:"value"`
}

// CollectionResourceDTO represents a collection resource
type CollectionResourceDTO struct {
	XMLName    xml.Name                         `xml:"resource"`
	Name       string                           `xml:"name,attr"`
	Resources  []CollectionResourceDTO          `xml:"resource"`
	Attributes []CollectionResourceAttributeDTO `xml:"attribute"`
}

// AddAttribute adds a new attribute to the resource
func (resource *CollectionResourceDTO) AddAttribute(aType string, aName string, aContent string) {
	attr := CollectionResourceAttributeDTO{
		Name: aName,
		Value: &CollectionAttributeValueDTO{
			Type:    aType,
			Content: aContent,
		},
	}
	resource.Attributes = append(resource.Attributes, attr)
}

// CollectionSetDTO represents a collection set
type CollectionSetDTO struct {
	XMLName                   xml.Name                `xml:"collection-set"`
	Timestamp                 *Timestamp              `xml:"timestamp,attr"`
	Status                    string                  `xml:"collection-status,attr"`
	DisableCounterPersistence bool                    `xml:"disable-counter-persistence,attr"`
	SequenceNumber            int                     `xml:"sequence-number,attr"`
	Agent                     *CollectionAgentDTO     `xml:"agent"`
	Resources                 []CollectionResourceDTO `xml:"collection-resource"`
}

// AddResource adds a new resource to the collection set
func (set *CollectionSetDTO) AddResource(resource CollectionResourceDTO) {
	set.Resources = append(set.Resources, resource)
}

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
	ForeignSource       string                   `xml:"foreign-source,attr,omitempty"`
	ForeignID           string                   `xml:"foreign-id,attr,omitempty"`
	Location            string                   `xml:"location,attr,omitempty"`
	StorageResourcePath string                   `xml:"storage-resource-path,attr"`
	SysUpTime           int64                    `xml:"sys-up-time,attr"`
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

// GetAttributeValue gets the value of a given attribute
func (req *CollectorRequestDTO) GetAttributeValue(key string, defaultValue string) string {
	for _, attr := range req.Attributes {
		if attr.Key == key {
			s := strings.Replace(attr.Content, "<![CDATA[", "", -1)
			return strings.Replace(s, "]]>", "", -1)
		}
	}
	return defaultValue
}

// CollectorResponseDTO represents a collector response
type CollectorResponseDTO struct {
	XMLName       xml.Name          `xml:"collector-response"`
	Error         string            `xml:"error,attr,omitempty"`
	CollectionSet *CollectionSetDTO `xml:"collection-set"`
}

// GetStatus returns the collection status as a string
func (set *CollectorResponseDTO) GetStatus() string {
	if set.CollectionSet == nil {
		return "nothing collected"
	}
	total := 0
	for _, r := range set.CollectionSet.Resources {
		total += len(r.Attributes)
	}
	return fmt.Sprintf("%d attributes in %d resources", total, len(set.CollectionSet.Resources))
}

// RRA represents an RRA object
type RRA struct {
	XMLName xml.Name `xml:"rra"`
	Content string   `xml:",chardata"`
}

// RRD represents an RRD object
type RRD struct {
	XMLName xml.Name `xml:"rrd"`
	Step    int      `xml:"step,attr"`
	RRAs    []RRA    `xml:"rra"`
}
