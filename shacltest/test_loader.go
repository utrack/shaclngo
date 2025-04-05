package shacltest

import (
	"fmt"

	"github.com/deiu/rdf2go"
	"github.com/utrack/shaclngo/rdft"
)

// Namespaces used in the SHACL test suite
const (
	RDF      = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	RDFS     = "http://www.w3.org/2000/01/rdf-schema#"
	SHACL    = "http://www.w3.org/ns/shacl#"
	SHACLT   = "http://www.w3.org/ns/shacl-test#"
	MANIFEST = "http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#"
)

// Test represents a SHACL validation test
type Test struct {
	// ID is the URI of the test
	ID string `rdf:"@id"`

	// Description is the human-readable description of the test
	Description string `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`

	// Status is the status of the test (e.g., Approved)
	Status rdft.Resource `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#status"`

	// Action contains the test data and shapes
	Action *Action `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#action"`

	// ExpectedResult contains the expected validation report
	ExpectedResult *ValidationReport `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#result"`
}

// Action represents the test data and shapes
type Action struct {
	// ID is the URI of the action
	ID string `rdf:"@id"`

	// Data is the URI of the data graph
	Data string `rdf:"http://www.w3.org/ns/shacl-test#dataGraph"`

	// Shapes is the URI of the shapes graph
	Shapes string `rdf:"http://www.w3.org/ns/shacl-test#shapesGraph"`
}

// ValidationReport represents a SHACL validation report
type ValidationReport struct {
	// ID is the URI of the validation report
	ID string `rdf:"@id" rdfType:"http://www.w3.org/ns/shacl#ValidationReport"`

	// Type is the RDF type of the validation report
	Type rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`

	// Conforms indicates whether the data graph conforms to the shapes
	Conforms bool `rdf:"http://www.w3.org/ns/shacl#conforms"`

	// Results contains the validation results
	Results []*ValidationResult `rdf:"http://www.w3.org/ns/shacl#result"`
}

// ValidationResult represents a SHACL validation result
type ValidationResult struct {
	// ID is the URI of the validation result
	ID string `rdf:"@id" rdfType:"http://www.w3.org/ns/shacl#ValidationResult"`

	// FocusNode is the node that was validated
	FocusNode string `rdf:"http://www.w3.org/ns/shacl#focusNode"`

	// ResultSeverity is the severity of the validation result
	ResultSeverity rdft.Resource `rdf:"http://www.w3.org/ns/shacl#resultSeverity"`

	// SourceConstraintComponent is the constraint component that generated the result
	SourceConstraintComponent rdft.Resource `rdf:"http://www.w3.org/ns/shacl#sourceConstraintComponent"`

	// SourceShape is the shape that generated the result
	SourceShape string `rdf:"http://www.w3.org/ns/shacl#sourceShape"`

	// Value is the value that failed validation
	Value string `rdf:"http://www.w3.org/ns/shacl#value"`
}

// GetTestManifests retrieves all SHACL validation tests from the given graph
func GetTestManifests(g *rdf2go.Graph) ([]Test, error) {
	// Create an unmarshaller for the graph
	unmarshaller := rdft.NewUnmarshaller(g)

	// Find all tests of type Validate
	testTriples := g.All(
		nil,
		rdf2go.NewResource(RDF+"type"),
		rdf2go.NewResource(SHACLT+"Validate"),
	)

	// Create a slice to hold the tests
	tests := make([]Test, 0, len(testTriples))

	// Unmarshal each test
	for _, triple := range testTriples {
		// Get the test URI
		testURI := triple.Subject.RawValue()

		// Create a new test
		var test Test

		// Unmarshal the test
		err := unmarshaller.Unmarshal(testURI, &test)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal test %s: %v", testURI, err)
		}

		// Add the test to the slice
		tests = append(tests, test)
	}

	return tests, nil
}
