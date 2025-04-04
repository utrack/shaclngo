// Package rdft provides functionality for unmarshalling RDF triples into Go structs.
package rdft

import (
	"fmt"

	"github.com/deiu/rdf2go"
)

// Value represents any RDF value (Resource, Literal, BlankNode)
type Value interface {
	// RawValue returns the raw string value
	RawValue() string

	// Type returns the RDF type of the value
	Type() string
}

// Resource represents an RDF resource (URI)
type Resource struct {
	URI string
}

// NewResource creates a new resource from a URI
func NewResource(uri string) *Resource {
	return &Resource{URI: uri}
}

// RawValue returns the raw URI value
func (r *Resource) RawValue() string {
	return r.URI
}

// Type returns the RDF type of the value
func (r *Resource) Type() string {
	return "resource"
}

// String returns a string representation of the resource
func (r *Resource) String() string {
	return fmt.Sprintf("<%s>", r.URI)
}

// Literal represents an RDF literal value
type Literal struct {
	Value    string
	Language string
	Datatype string
}

// NewLiteral creates a new literal with the given value
func NewLiteral(value string) *Literal {
	return &Literal{Value: value}
}

// NewLiteralWithLanguage creates a new literal with the given value and language
func NewLiteralWithLanguage(value, language string) *Literal {
	return &Literal{Value: value, Language: language}
}

// NewLiteralWithDatatype creates a new literal with the given value and datatype
func NewLiteralWithDatatype(value, datatype string) *Literal {
	return &Literal{Value: value, Datatype: datatype}
}

// RawValue returns the raw string value
func (l *Literal) RawValue() string {
	return l.Value
}

// Type returns the RDF type of the value
func (l *Literal) Type() string {
	if l.Datatype != "" {
		return l.Datatype
	}
	if l.Language != "" {
		return "langString"
	}
	return "literal"
}

// String returns a string representation of the literal
func (l *Literal) String() string {
	if l.Language != "" {
		return fmt.Sprintf("\"%s\"@%s", l.Value, l.Language)
	}
	if l.Datatype != "" {
		return fmt.Sprintf("\"%s\"^^<%s>", l.Value, l.Datatype)
	}
	return fmt.Sprintf("\"%s\"", l.Value)
}

// BlankNode represents an RDF blank node
type BlankNode struct {
	ID string
}

// NewBlankNode creates a new blank node with the given ID
func NewBlankNode(id string) *BlankNode {
	return &BlankNode{ID: id}
}

// RawValue returns the raw ID value
func (b *BlankNode) RawValue() string {
	return b.ID
}

// Type returns the RDF type of the value
func (b *BlankNode) Type() string {
	return "blankNode"
}

// String returns a string representation of the blank node
func (b *BlankNode) String() string {
	return fmt.Sprintf("_:%s", b.ID)
}

// LocalizedString represents a string with a language tag
type LocalizedString struct {
	Value    string
	Language string
}

// NewLocalizedString creates a new localized string
func NewLocalizedString(value, language string) *LocalizedString {
	return &LocalizedString{
		Value:    value,
		Language: language,
	}
}

// String returns a string representation of the localized string
func (ls *LocalizedString) String() string {
	if ls.Language == "" {
		return ls.Value
	}
	return fmt.Sprintf("%s@%s", ls.Value, ls.Language)
}

// LocalizedText represents a collection of translations for the same text
type LocalizedText struct {
	Translations map[string]string // language code -> text
}

// NewLocalizedText creates a new empty localized text collection
func NewLocalizedText() *LocalizedText {
	return &LocalizedText{
		Translations: make(map[string]string),
	}
}

// Add adds or updates a translation
func (lt *LocalizedText) Add(language, text string) {
	lt.Translations[language] = text
}

// Get retrieves a translation by language code
func (lt *LocalizedText) Get(language string) (string, bool) {
	text, exists := lt.Translations[language]
	return text, exists
}

// GetWithFallback retrieves a translation, falling back to a default language if not found
func (lt *LocalizedText) GetWithFallback(language, fallback string) string {
	if text, exists := lt.Get(language); exists {
		return text
	}
	if text, exists := lt.Get(fallback); exists {
		return text
	}
	// Return any translation if nothing else is available
	for _, text := range lt.Translations {
		return text
	}
	return ""
}

// RDFUnmarshaler is an interface for types that can unmarshal themselves from RDF values
type RDFUnmarshaler interface {
	UnmarshalRDF(values []Value) error
}

// Convert RDF2Go term to rdft Value
func FromRDF2GoTerm(term rdf2go.Term) Value {
	switch t := term.(type) {
	case *rdf2go.Resource:
		return NewResource(t.URI)
	case *rdf2go.Literal:
		if t.Language != "" {
			return NewLiteralWithLanguage(t.Value, t.Language)
		}
		if t.Datatype != nil {
			dt, ok := t.Datatype.(*rdf2go.Resource)
			if ok {
				return NewLiteralWithDatatype(t.Value, dt.URI)
			}
		}
		return NewLiteral(t.Value)
	case *rdf2go.BlankNode:
		return NewBlankNode(t.ID)
	default:
		return nil
	}
}
