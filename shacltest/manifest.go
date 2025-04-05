package shacltest

import "github.com/utrack/shaclngo/rdft"

type Manifest struct {
	ID string `rdf:"@id" rdfType:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#Manifest"`

	// Entries contains the list of tests
	Entries []ValidationTest `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#entries"`
}

// ValidationTest represents a SHACL validation test
type ValidationTest struct {
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
	Data rdft.Resource `rdf:"http://www.w3.org/ns/shacl-test#dataGraph"`

	// Shapes is the URI of the shapes graph
	Shapes rdft.Resource `rdf:"http://www.w3.org/ns/shacl-test#shapesGraph"`
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
	FocusNode rdft.Resource `rdf:"http://www.w3.org/ns/shacl#focusNode"`

	// ResultSeverity is the severity of the validation result
	ResultSeverity rdft.Resource `rdf:"http://www.w3.org/ns/shacl#resultSeverity"`

	// SourceConstraintComponent is the constraint component that generated the result
	SourceConstraintComponent rdft.Resource `rdf:"http://www.w3.org/ns/shacl#sourceConstraintComponent"`

	// SourceShape is the shape that generated the result
	SourceShape rdft.Resource `rdf:"http://www.w3.org/ns/shacl#sourceShape"`

	// Value is the value that failed validation
	Value rdft.Resource `rdf:"http://www.w3.org/ns/shacl#value"`
}
