package rdft

import (
	"testing"

	"github.com/deiu/rdf2go"
	"github.com/stretchr/testify/assert"
)

// TestStructWithTypeAssertion defines a struct with a type assertion
type TestStructWithTypeAssertion struct {
	URI     string `rdf:"@id" rdfType:"http://example.org/Person"`
	Name    string `rdf:"http://xmlns.com/foaf/0.1/name"`
	Age     int    `rdf:"http://xmlns.com/foaf/0.1/age"`
}

// TestStructWithoutTypeAssertion defines a struct without a type assertion
type TestStructWithoutTypeAssertion struct {
	URI     string `rdf:"@id"`
	Name    string `rdf:"http://xmlns.com/foaf/0.1/name"`
	Age     int    `rdf:"http://xmlns.com/foaf/0.1/age"`
}

// TestTypeAssertion tests that type assertions work correctly
func TestTypeAssertion(t *testing.T) {
	// Create a new graph
	graph := rdf2go.NewGraph("http://example.org/")

	// Add triples for a person
	personURI := "http://example.org/person1"
	
	// Add type triple
	graph.AddTriple(
		rdf2go.NewResource(personURI),
		rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
		rdf2go.NewResource("http://example.org/Person"),
	)
	
	// Add name and age
	graph.AddTriple(
		rdf2go.NewResource(personURI),
		rdf2go.NewResource("http://xmlns.com/foaf/0.1/name"),
		rdf2go.NewLiteral("John Doe"),
	)
	graph.AddTriple(
		rdf2go.NewResource(personURI),
		rdf2go.NewResource("http://xmlns.com/foaf/0.1/age"),
		rdf2go.NewLiteral("30"),
	)

	// Create a new unmarshaller
	unmarshaller := NewUnmarshaller(graph)

	// Test 1: Successful type assertion
	t.Run("SuccessfulTypeAssertion", func(t *testing.T) {
		var person TestStructWithTypeAssertion
		err := unmarshaller.Unmarshal(personURI, &person)
		assert.NoError(t, err)
		assert.Equal(t, personURI, person.URI)
		assert.Equal(t, "John Doe", person.Name)
		assert.Equal(t, 30, person.Age)
	})

	// Test 2: No type assertion (should work)
	t.Run("NoTypeAssertion", func(t *testing.T) {
		var person TestStructWithoutTypeAssertion
		err := unmarshaller.Unmarshal(personURI, &person)
		assert.NoError(t, err)
		assert.Equal(t, personURI, person.URI)
		assert.Equal(t, "John Doe", person.Name)
		assert.Equal(t, 30, person.Age)
	})

	// Test 3: Failed type assertion
	t.Run("FailedTypeAssertion", func(t *testing.T) {
		// Create a struct with a different expected type
		type TestStructWithWrongTypeAssertion struct {
			URI  string `rdf:"@id" rdfType:"http://example.org/Organization"`
			Name string `rdf:"http://xmlns.com/foaf/0.1/name"`
		}

		var org TestStructWithWrongTypeAssertion
		err := unmarshaller.Unmarshal(personURI, &org)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not have expected RDF type")
	})

	// Test 4: Missing type triple
	t.Run("MissingTypeTriple", func(t *testing.T) {
		// Create a resource without a type triple
		noTypeURI := "http://example.org/person2"
		graph.AddTriple(
			rdf2go.NewResource(noTypeURI),
			rdf2go.NewResource("http://xmlns.com/foaf/0.1/name"),
			rdf2go.NewLiteral("Jane Doe"),
		)

		var person TestStructWithTypeAssertion
		err := unmarshaller.Unmarshal(noTypeURI, &person)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has no RDF type")
	})
}
