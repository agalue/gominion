package collectors

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/tools"
)

const xmlCollectionAttr = "xmlDatacollection"

// XMLCollector represents a collector implementation
type XMLCollector struct {
}

// GetID gets the collector ID (simple class name from its Java counterpart)
func (collector *XMLCollector) GetID() string {
	return "XmlCollector"
}

// Collect execute the collector request and return the collection response
func (collector *XMLCollector) Collect(request *api.CollectorRequestDTO) *api.CollectorResponseDTO {
	response := &api.CollectorResponseDTO{}
	xmlCollection := &api.XMLCollection{}
	err := xml.Unmarshal([]byte(request.GetAttributeValue(xmlCollectionAttr, "")), xmlCollection)
	if err != nil {
		response.MarkAsFailed(request.CollectionAgent, fmt.Errorf("Cannot parse %s: %v", xmlCollectionAttr, err))
		return response
	}
	builder := api.NewCollectionSetBuilder(request.CollectionAgent)
	handlerClass := request.GetAttributeValue("handler-class", XMLHandlerClass)
	for _, src := range xmlCollection.Sources {
		querier, err := NewQuerier(handlerClass, src.GetRequest())
		if err != nil {
			response.MarkAsFailed(request.CollectionAgent, err)
			return response
		}
		log.Debugf("Executing an HTTP GET against %s", src.URL)
		if doc, err := collector.getDocument(querier, src, request.GetTimeout()); err != nil {
			response.MarkAsFailed(request.CollectionAgent, err)
			return response
		} else if err := collector.fillCollectionSet(querier, src, builder, doc); err != nil {
			response.MarkAsFailed(request.CollectionAgent, err)
			return response
		}
	}
	response.CollectionSet = builder.Build()
	return response
}

func (collector *XMLCollector) fillCollectionSet(querier XPathQuerier, src api.XMLSource, builder *api.CollectionSetBuilder, doc *XPathNode) error {
	re, _ := regexp.Compile("[.\\d]+")
	for _, group := range src.Groups {
		resources, err := querier.QueryAll(doc, group.ResourceXPath)
		if err != nil {
			return err
		}
		timestamp := collector.getTimestamp(group, doc)
		for _, resource := range resources {
			name, err := collector.getResourceName(querier, group, resource)
			if err != nil {
				return err
			}
			cres := collector.getCollectionResource(builder.Agent, name, group.ResourceType, timestamp)
			for _, obj := range group.Objects {
				value, err := querier.Query(resource, obj.XPath)
				if err != nil {
					return err
				}
				v := value.GetContent()
				if obj.Type != "string" {
					data := re.FindAllString(value.GetContent(), -1)
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

func (collector *XMLCollector) getResourceName(querier XPathQuerier, group api.XMLGroup, node *XPathNode) (string, error) {
	if group.HasMultipleResourceKeys() {
		keys := make([]string, 0)
		for _, key := range group.ResourceKey.KeyXPaths {
			keyNode, err := querier.Query(node, key)
			if err != nil {
				keys = append(keys, keyNode.GetContent())
			}
			return "", err
		}
		return strings.Join(keys, "_"), nil
	}
	if group.KeyXPath == "" {
		log.Debugf("Assuming node level resource")
		return "node", nil
	}
	keyNode, err := querier.Query(node, group.KeyXPath)
	if err == nil {
		return keyNode.GetContent(), nil
	}
	return "", fmt.Errorf("Cannot find resource name")
}

// TODO pending parse for group.TimestampXPath and group.TimestampFormat
func (collector *XMLCollector) getTimestamp(group api.XMLGroup, node *XPathNode) *api.Timestamp {
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

func (collector *XMLCollector) getDocument(querier XPathQuerier, src api.XMLSource, timeout time.Duration) (*XPathNode, error) {
	httpreq, err := src.GetHTTPRequest()
	if err != nil {
		return nil, err
	}
	if t := src.GetRequest().GetParameterAsInt("timeout"); t > 0 {
		timeout = time.Duration(t) * time.Microsecond
	}
	client := tools.GetHTTPClient(src.SkipSSL(), timeout)
	httpres, err := client.Do(httpreq)
	if err != nil {
		return nil, err
	}
	doc, err := querier.Parse(httpres.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

var xmlCollector = &XMLCollector{}

func init() {
	RegisterCollector(xmlCollector)
}
