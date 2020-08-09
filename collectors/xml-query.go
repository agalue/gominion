package collectors

import (
	"fmt"
	"io"

	"github.com/agalue/gominion/api"
	"github.com/andybalholm/cascadia"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xmlquery"
	"golang.org/x/net/html"
)

const (
	// XMLHandlerClass represents the Java class-name of the default XML collection handler
	XMLHandlerClass = "org.opennms.protocols.xml.collector.DefaultXmlCollectionHandler"

	// JSONHandlerClass represents the Java class-name of the default JSON collection handler
	JSONHandlerClass = "org.opennms.protocols.json.collector.DefaultJsonCollectionHandler"

	// HTTPHandlerClass represents the Java class-name of the default HTML/CSS collection handler
	HTTPHandlerClass = "org.opennms.protocols.json.collector.HttpCollectionHandler"
)

// XPathQuerier an interface that represents an XPath queries
type XPathQuerier interface {
	Parse(reader io.Reader) (*XPathNode, error)
	Query(parent *XPathNode, xpath string) (*XPathNode, error)
	QueryAll(parent *XPathNode, xpath string) ([]*XPathNode, error)
}

// XPathNode represents a node
type XPathNode struct {
	impl interface{}
}

// GetContent gets the content of a given node
func (n *XPathNode) GetContent() string {
	switch o := n.impl.(type) {
	case *xmlquery.Node:
		return o.InnerText()
	case *jsonquery.Node:
		if o == nil {
			return ""
		}
		return o.InnerText()
	case *html.Node:
		if o == nil {
			return ""
		}
		return htmlquery.InnerText(o)
	default:
		return ""
	}
}

// XPathQuery represents a query
type XPathQuery struct {
	kind string
}

// Parse parses the content from a reader and return the root node
func (q *XPathQuery) Parse(reader io.Reader) (*XPathNode, error) {
	var err error
	var node interface{}
	switch q.kind {
	case "css":
		node, err = html.Parse(reader)
	case "xml":
		node, err = xmlquery.Parse(reader)
	case "html":
		node, err = htmlquery.Parse(reader)
	case "json":
		node, err = jsonquery.Parse(reader)
	default:
		return nil, fmt.Errorf("Cannot find implementation")
	}
	if err == nil {
		return &XPathNode{impl: node}, nil
	}
	return nil, err
}

// Query queries the parent node and returns the first match
func (q *XPathQuery) Query(parent *XPathNode, xpath string) (*XPathNode, error) {
	var err error = nil
	var node interface{} = nil
	switch q.kind {
	case "css":
		p := parent.impl.(*html.Node)
		sel, err := cascadia.Compile(xpath)
		if err == nil {
			node = sel.MatchFirst(p)
		}
	case "xml":
		p := parent.impl.(*xmlquery.Node)
		node, err = xmlquery.Query(p, xpath)
	case "html":
		p := parent.impl.(*html.Node)
		node, err = htmlquery.Query(p, xpath)
	case "json":
		p := parent.impl.(*jsonquery.Node)
		node, err = jsonquery.Query(p, xpath)
	default:
		return nil, fmt.Errorf("Cannot find implementation")
	}
	if err == nil {
		if node == nil {
			return nil, fmt.Errorf("Cannot find element")
		}
		return &XPathNode{impl: node}, nil
	}
	return nil, err
}

// QueryAll queries the parent node and returns an array with all the matches
func (q *XPathQuery) QueryAll(parent *XPathNode, xpath string) ([]*XPathNode, error) {
	switch q.kind {
	case "css":
		p := parent.impl.(*html.Node)
		sel, err := cascadia.Compile(xpath)
		if err != nil {
			return nil, err
		}
		list := sel.MatchAll(p)
		nodes := make([]*XPathNode, len(list))
		for i, n := range list {
			nodes[i] = &XPathNode{impl: n}
		}
		return nodes, nil
	case "xml":
		p := parent.impl.(*xmlquery.Node)
		list, err := xmlquery.QueryAll(p, xpath)
		if err != nil {
			return nil, err
		}
		nodes := make([]*XPathNode, len(list))
		for i, n := range list {
			nodes[i] = &XPathNode{impl: n}
		}
		return nodes, nil
	case "html":
		p := parent.impl.(*html.Node)
		list, err := htmlquery.QueryAll(p, xpath)
		if err != nil {
			return nil, err
		}
		nodes := make([]*XPathNode, len(list))
		for i, n := range list {
			nodes[i] = &XPathNode{impl: n}
		}
		return nodes, nil
	case "json":
		p := parent.impl.(*jsonquery.Node)
		list, err := jsonquery.QueryAll(p, xpath)
		if err != nil {
			return nil, err
		}
		nodes := make([]*XPathNode, len(list))
		for i, n := range list {
			nodes[i] = &XPathNode{impl: n}
		}
		return nodes, nil
	default:
		return nil, fmt.Errorf("Cannot find implementation")
	}
}

// NewQuerier returns new querier interface for a given kind
func NewQuerier(handlerClass string, req *api.XMLRequest) (XPathQuerier, error) {
	switch handlerClass {
	case XMLHandlerClass:
		return &XPathQuery{kind: "xml"}, nil
	case JSONHandlerClass:
		return &XPathQuery{kind: "json"}, nil
	case HTTPHandlerClass:
		return &XPathQuery{kind: "css"}, nil
	}
	if req != nil && req.GetParameterAsString("pre-parse-html") == "true" {
		return &XPathQuery{kind: "html"}, nil
	}
	return nil, fmt.Errorf("Cannot find suitable implementation for %s", handlerClass)
}
