package shacl

import "github.com/utrack/shaclngo/rdft"

type Shape struct {
	ID rdft.Resource `rdf:"@id" rdfType:"http://www.w3.org/ns/shacl#Shape"`

	Deactivated *bool `rdf:"http://www.w3.org/ns/shacl#deactivated"`

	Message *rdft.LocalizedString `rdf:"http://www.w3.org/ns/shacl#message"`

	Severity *rdft.Resource `rdf:"http://www.w3.org/ns/shacl#severity"`

	Target *Target `rdf:"http://www.w3.org/ns/shacl#target"`
}

type PropertyShape struct {
	Shape

	ID   rdft.Resource         `rdf:"@id" rdfType:"http://www.w3.org/ns/shacl#PropertyShape"`
	Name *rdft.LocalizedString `rdf:"http://www.w3.org/ns/shacl#name"`
}

type NodeShape struct {
	Shape
	ID rdft.Resource `rdf:"@id" rdfType:"http://www.w3.org/ns/shacl#NodeShape"`
}

type Target struct {
	Class      []rdft.Resource `rdf:"http://www.w3.org/ns/shacl#targetClass"`
	Node       []rdft.Resource `rdf:"http://www.w3.org/ns/shacl#targetNode"`
	SubjectsOf []rdft.Resource `rdf:"http://www.w3.org/ns/shacl#targetSubjectsOf"`
	ObjectsOf  []rdft.Resource `rdf:"http://www.w3.org/ns/shacl#targetObjectsOf"`
}

const (
	NodeKindBlankNode = "http://www.w3.org/ns/shacl#BlankNode"
	NodeKindIRI       = "http://www.w3.org/ns/shacl#IRI"
	NodeKindLiteral   = "http://www.w3.org/ns/shacl#Literal"
	NodeKindBlankNodeOrIRI = "http://www.w3.org/ns/shacl#BlankNodeOrIRI"
	NodeKindBlankNodeOrLiteral = "http://www.w3.org/ns/shacl#BlankNodeOrLiteral"
	NodeKindIRIOrLiteral = "http://www.w3.org/ns/shacl#IRIOrLiteral"
)