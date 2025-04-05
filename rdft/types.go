// Package rdft provides functionality for unmarshalling RDF triples into Go structs.
package rdft

import (
	"fmt"

	"github.com/deiu/rdf2go"
)

type Resource = rdf2go.Resource


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
	UnmarshalRDF(values []*rdf2go.Triple) error
}
