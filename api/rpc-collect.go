package api

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CollectionStatusUnknown status unknown
const CollectionStatusUnknown = "UNKNOWN"

// CollectionStatusSucceded collection finished successfully
const CollectionStatusSucceded = "SUCCEEDED"

// CollectionStatusFailed collection failed
const CollectionStatusFailed = "FAILED"

// ResourceAttributeDTO represents a generic resource attribute
type ResourceAttributeDTO struct {
	Group      string `xml:"group,attr"`
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
	Identifier string `xml:"identifier,attr,omitempty"`
	Value      string `xml:"value,attr"`
}

// GenericTypeResourceDTO represents a Generic Level Resource Type
type GenericTypeResourceDTO struct {
	XMLName   xml.Name              `xml:"generic-type-resource"`
	Node      *NodeLevelResourceDTO `xml:"node-level-resource"`
	Name      string                `xml:"name,attr"`
	Fallback  string                `xml:"fallback,attr"`
	Instance  string                `xml:"instance,attr"`
	Timestamp *Timestamp            `xml:"timestamp,attr,omitempty"`
}

// InterfaceLevelResourceDTO represents an Interface Level Resource Type
type InterfaceLevelResourceDTO struct {
	XMLName   xml.Name              `xml:"interface-level-resource"`
	Node      *NodeLevelResourceDTO `xml:"node-level-resource"`
	IntfName  string                `xml:"if-name,attr"`
	Timestamp *Timestamp            `xml:"timestamp,attr,omitempty"`
}

// NodeLevelResourceDTO represents a Node Level Resource Type
type NodeLevelResourceDTO struct {
	XMLName   xml.Name   `xml:"node-level-resource"`
	NodeID    int        `xml:"node-id,attr"`
	Path      string     `xml:"path,attr,omitempty"`
	Timestamp *Timestamp `xml:"timestamp,attr,omitempty"`
}

// CollectionResourceDTO represents a collection resource
type CollectionResourceDTO struct {
	XMLName           xml.Name                `xml:"collection-resource"`
	Name              string                  `xml:"name,attr,omitempty"`
	ResourceType      interface{}             // NodeLevelResourceDTO, InterfaceLevelResourceDTO, or GenericTypeResourceDTO
	Resources         []CollectionResourceDTO `xml:"resource"`
	NumericAttributes []ResourceAttributeDTO  `xml:"numeric-attribute"`
	StringAttributes  []ResourceAttributeDTO  `xml:"string-attribute"`
}

// AddNumericAttribute adds a new numeric attribute to the resource
func (resource *CollectionResourceDTO) AddNumericAttribute(name string, group string, id string, ntype string, value string) {
	attr := ResourceAttributeDTO{
		Name:       name,
		Group:      group,
		Value:      value,
		Type:       ntype,
		Identifier: id,
	}
	resource.NumericAttributes = append(resource.NumericAttributes, attr)
}

// AddStringAttribute adds a new string attribute to the resource
func (resource *CollectionResourceDTO) AddStringAttribute(name string, group string, id string, value string) {
	attr := ResourceAttributeDTO{
		Name:       name,
		Group:      group,
		Value:      value,
		Type:       "string",
		Identifier: id,
	}
	resource.StringAttributes = append(resource.StringAttributes, attr)
}

// CollectionSetDTO represents a collection set
type CollectionSetDTO struct {
	XMLName                   xml.Name                `xml:"collection-set"`
	Timestamp                 *Timestamp              `xml:"timestamp,attr"`
	Status                    string                  `xml:"collection-status,attr"`
	DisableCounterPersistence bool                    `xml:"disable-counter-persistence,attr,omitempty"`
	SequenceNumber            int                     `xml:"sequence-number,attr,omitempty"`
	Agent                     *CollectionAgentDTO     `xml:"agent"`
	Resources                 []CollectionResourceDTO `xml:"collection-resource"`
}

// AddResource adds a new resource to the collection set
func (set *CollectionSetDTO) AddResource(resource CollectionResourceDTO) {
	set.Resources = append(set.Resources, resource)
}

// CollectionAttributeDTO represents a collection attribute
type CollectionAttributeDTO struct {
	XMLName xml.Name `xml:"attribute"`
	Key     string   `xml:"key,attr"`
	Content string   `xml:",innerxml"`
}

// CollectionAgentDTO represents a collection agent
type CollectionAgentDTO struct {
	XMLName             xml.Name                 `xml:"agent"`
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

// GetResourcePath returns the base resource path
func (req *CollectorRequestDTO) GetResourcePath() string {
	a := req.CollectionAgent
	if !a.StoreByFS || a.ForeignSource == "" || a.ForeignID == "" {
		return fmt.Sprintf("snmp/%d", a.NodeID)
	}
	return fmt.Sprintf("snmp/fs/%s/%s", a.ForeignSource, a.ForeignID)
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

// GetTimeout extracts the duration of the timeout attribute if available; otherwise returns default value
func (req *CollectorRequestDTO) GetTimeout() time.Duration {
	if value := req.GetAttributeValue("timeout", ""); value != "" {
		if t, err := strconv.Atoi(value); err != nil {
			return time.Duration(t) * time.Millisecond
		}
	}
	return DefaultTimeout
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
		total += len(r.NumericAttributes)
		total += len(r.StringAttributes)
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
