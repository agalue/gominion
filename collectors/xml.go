package collectors

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/tools"
)

const xmlCollectionAttr = "xmlDatacollection"

// XMLCollector represents a collector implementation
type XMLCollector struct {
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (collector *XMLCollector) GetID() string {
	return "XmlCollector"
}

// Collect execute the collector request and return the collection set
// FIXME Assumes XML Handler. Needs code for managing JSON and HTML
func (collector *XMLCollector) Collect(request *api.CollectorRequestDTO) *api.CollectorResponseDTO {
	response := &api.CollectorResponseDTO{}
	xmlCollection := &api.XMLCollection{}
	err := xml.Unmarshal([]byte(request.GetAttributeValue(xmlCollectionAttr, "")), xmlCollection)
	if err != nil {
		response.MarkAsFailed(request.CollectionAgent, fmt.Errorf("Cannot parse %s: %v", xmlCollectionAttr, err))
		return response
	}
	builder := api.NewCollectionSetBuilder(request.CollectionAgent)
	for _, src := range xmlCollection.Sources {
		log.Debugf("Executing an HTTP GET against %s", src.URL)
		if doc, err := collector.getDocument(&src, request.GetTimeout()); err != nil {
			response.MarkAsFailed(request.CollectionAgent, err)
			return response
		} else if err := collector.fillCollectionSet(builder, src, doc); err != nil {
			response.MarkAsFailed(request.CollectionAgent, err)
			return response
		}
	}
	response.CollectionSet = builder.Build()
	return response
}

func (collector *XMLCollector) fillCollectionSet(builder *api.CollectionSetBuilder, src api.XMLSource, doc *xmlquery.Node) error {
	re, _ := regexp.Compile("[.\\d]+")
	for _, group := range src.Groups {
		resources, err := xmlquery.QueryAll(doc, group.ResourceXPath)
		if err != nil {
			return err
		}
		timestamp := collector.getTimestamp(group, doc)
		for _, resource := range resources {
			name, err := collector.getResourceName(group, resource)
			if err != nil {
				return err
			}
			cres := collector.getCollectionResource(builder.Agent, name, group.ResourceType, timestamp)
			for _, obj := range group.Objects {
				value, err := xmlquery.Query(resource, obj.XPath)
				if err != nil {
					return err
				}
				v := value.InnerText()
				if obj.Type != "string" {
					data := re.FindAllString(value.InnerText(), -1)
					if len(data) > 0 {
						v = data[0]
					}
				}
				builder.WithAttribute(cres, group.Name, obj.Name, v, obj.Type)
			}
		}
	}
	return nil
}

func (collector *XMLCollector) getResourceName(group api.XMLGroup, node *xmlquery.Node) (string, error) {
	if group.HasMultipleResourceKeys() {
		keys := make([]string, 0)
		for _, key := range group.ResourceKey.KeyXPaths {
			keyNode, err := xmlquery.Query(node, key)
			if err != nil {
				keys = append(keys, keyNode.InnerText())
			}
			return "", err
		}
		return strings.Join(keys, "_"), nil
	}
	if group.KeyXPath == "" {
		log.Debugf("Assuming node level resource")
		return "node", nil
	}
	keyNode, err := xmlquery.Query(node, group.KeyXPath)
	if err == nil {
		return keyNode.InnerText(), nil
	}
	return "", fmt.Errorf("Cannot find resource name")
}

// TODO pending parse for group.TimestampXPath and group.TimestampFormat
func (collector *XMLCollector) getTimestamp(group api.XMLGroup, node *xmlquery.Node) *api.Timestamp {
	return &api.Timestamp{Time: time.Now()}
}

func (collector *XMLCollector) getCollectionResource(agent *api.CollectionAgentDTO, instance string, resourceType string, timestamp *api.Timestamp) *api.CollectionResourceDTO {
	nodeType := &api.NodeLevelResourceDTO{
		NodeID: agent.NodeID,
	}
	if resourceType == "node" {
		return &api.CollectionResourceDTO{ResourceType: nodeType}
	}
	return &api.CollectionResourceDTO{
		ResourceType: &api.GenericTypeResourceDTO{
			Node:      nodeType,
			Name:      resourceType,
			Timestamp: timestamp,
			Instance:  instance,
		},
	}
}

func (collector *XMLCollector) getDocument(src *api.XMLSource, timeout time.Duration) (*xmlquery.Node, error) {
	httpreq, err := src.GetHTTPRequest()
	if err != nil {
		return nil, err
	}
	client := tools.GetHTTPClient(false, timeout)
	httpres, err := client.Do(httpreq)
	if err != nil {
		return nil, err
	}
	doc, err := xmlquery.Parse(httpres.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

var xmlCollector = &XMLCollector{}

func init() {
	RegisterCollector(xmlCollector)
}
