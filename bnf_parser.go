package gqlgen

import (
	"encoding/xml"
	"io"
	"log"
	"reflect"
)

type bnfParser struct {
}

func newBNFParser() bnfParser {
	return bnfParser{}
}

func (p *bnfParser) parse(r io.Reader) *node {
	decoder := xml.NewDecoder(r)
	return p.buildTree(decoder)
}

func (p *bnfParser) buildTree(decoder *xml.Decoder) *node {
	var root, current *node
	for token, err := decoder.Token(); err == nil; token, err = decoder.Token() {
		switch v := token.(type) {
		case xml.StartElement:
			current = p.handleStartElement(v, current)
			if root == nil {
				root = current
			}
		case xml.EndElement:
			current = current.parent
		case xml.ProcInst:
		case xml.CharData:
			if current != nil {
				switch current.id {
				case kwId:
					current.value = string(v)
				case terminalSymbolId:
					current.value = string(v)
				default:
				}
			}
		case xml.Comment:
		case xml.Directive:
		default:
			log.Fatalf("from buildGraph'case %s:' not handled\n", reflect.TypeOf(v))
		}
	}
	return root
}

func (b *bnfParser) handleStartElement(start xml.StartElement, parent *node) *node {
	n := &node{
		id:       start.Name.Local,
		parent:   parent,
		children: make([]*node, 0),
	}
	if n.id == bnfDefId || n.id == bnfId {
		n.name = b.attrMap(start.Attr)["name"]
	}
	if parent != nil {
		parent.children = append(parent.children, n)
	}
	return n
}

func (b *bnfParser) attrMap(attrs []xml.Attr) map[string]string {
	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Name.Local] = attr.Value
	}
	return attrMap
}
