package rdft

import (
	"testing"

	"github.com/deiu/rdf2go"
	"github.com/stretchr/testify/assert"
)

// TestStructWithLocalizedFields defines a struct with localized fields
// without the explicit rdf-type:"localized" tag
type TestStructWithLocalizedFields struct {
	// No explicit rdf-type:"localized" tag
	Title       LocalizedString `rdf:"http://example.org/title"`
	Description *LocalizedText  `rdf:"http://example.org/description"`
}

// TestAutoLocalizedFields tests that LocalizedString and LocalizedText fields
// are automatically treated as localized without requiring the rdf-type:"localized" tag
func TestAutoLocalizedFields(t *testing.T) {
	// Create a new graph
	graph := rdf2go.NewGraph("http://example.org/")

	// Add triples for a resource with localized text
	subject := "http://example.org/resource1"
	
	// Add title in English
	graph.AddTriple(
		rdf2go.NewResource(subject),
		rdf2go.NewResource("http://example.org/title"),
		rdf2go.NewLiteralWithLanguage("Title in English", "en"),
	)

	// Add description in multiple languages
	graph.AddTriple(
		rdf2go.NewResource(subject),
		rdf2go.NewResource("http://example.org/description"),
		rdf2go.NewLiteralWithLanguage("Description in English", "en"),
	)
	graph.AddTriple(
		rdf2go.NewResource(subject),
		rdf2go.NewResource("http://example.org/description"),
		rdf2go.NewLiteralWithLanguage("Description in French", "fr"),
	)
	graph.AddTriple(
		rdf2go.NewResource(subject),
		rdf2go.NewResource("http://example.org/description"),
		rdf2go.NewLiteralWithLanguage("Description in German", "de"),
	)

	// Create a new unmarshaller
	unmarshaller := NewUnmarshaller(graph)

	// Unmarshal the resource
	var result TestStructWithLocalizedFields
	err := unmarshaller.Unmarshal(subject, &result)
	assert.NoError(t, err)

	// Check that the title was correctly unmarshalled
	assert.Equal(t, "Title in English", result.Title.Value)
	assert.Equal(t, "en", result.Title.Language)

	// Check that the description was correctly unmarshalled
	assert.NotNil(t, result.Description)
	
	// Check each language
	enText, enOk := result.Description.Get("en")
	assert.True(t, enOk)
	assert.Equal(t, "Description in English", enText)
	
	frText, frOk := result.Description.Get("fr")
	assert.True(t, frOk)
	assert.Equal(t, "Description in French", frText)
	
	deText, deOk := result.Description.Get("de")
	assert.True(t, deOk)
	assert.Equal(t, "Description in German", deText)
}
