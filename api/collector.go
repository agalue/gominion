package api

import (
	"encoding/xml"
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
	Value   *CollectionAttributeValueDTO `xml:"value,attr"`
}

// CollectionResourceDTO represents a collection resource
type CollectionResourceDTO struct {
	XMLName    xml.Name                 `xml:"resource"`
	Name       string                   `xml:"name,attr"`
	Resources  []CollectionResourceDTO  `xml:"resource"`
	Attributes []CollectionAttributeDTO `xml:"attribute"`
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

// ServiceCollector represents the service collector interface
type ServiceCollector interface {
	GetID() string
	Collect(request *CollectorRequestDTO) CollectionSetDTO
}
