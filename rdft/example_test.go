package rdft_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/deiu/rdf2go"
	"github.com/utrack/shaclngo/rdft"
)

// Example struct representing a Rectangle from the SHACL test
type Rectangle struct {
	URI    rdft.Resource `rdf:"@id"`
	Type   rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
	Width  int           `rdf:"http://datashapes.org/sh/tests/core/node/and-001.test#width"`
	Height int           `rdf:"http://datashapes.org/sh/tests/core/node/and-001.test#height"`
}

// Example struct representing a ValidationResult from the SHACL test
type ValidationResult struct {
	URI                       rdft.Resource `rdf:"@id"`
	Type                      rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
	FocusNode                 rdft.Resource `rdf:"http://www.w3.org/ns/shacl#focusNode"`
	ResultSeverity            rdft.Resource `rdf:"http://www.w3.org/ns/shacl#resultSeverity"`
	SourceConstraintComponent rdft.Resource `rdf:"http://www.w3.org/ns/shacl#sourceConstraintComponent"`
	SourceShape               rdft.Resource `rdf:"http://www.w3.org/ns/shacl#sourceShape"`
	Value                     rdft.Resource `rdf:"http://www.w3.org/ns/shacl#value"`
}

// Example struct representing a ValidationReport from the SHACL test
type ValidationReport struct {
	URI      rdft.Resource      `rdf:"@id"`
	Type     rdft.Resource      `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
	Conforms bool               `rdf:"http://www.w3.org/ns/shacl#conforms"`
	Results  []ValidationResult `rdf:"http://www.w3.org/ns/shacl#result"`
}

// Example struct representing a TestCase from the SHACL test
type TestCase struct {
	URI    rdft.Resource     `rdf:"@id"`
	Type   rdft.Resource     `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
	Label  string            `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
	Action map[string]string `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#action"`
	Result *ValidationReport `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#result"`
	Status rdft.Resource     `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#status"`
}

// Example struct representing a Manifest from the SHACL test
type Manifest struct {
	URI     rdft.Resource `rdf:"@id"`
	Type    rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
	Entries []TestCase    `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#entries"`
}

// loadGraph loads an RDF graph from a Turtle file
func loadGraph(path string) (*rdf2go.Graph, error) {
	// Use a fixed base URI to match the one in the test file
	g := rdf2go.NewGraph("http://datashapes.org/sh/tests/core/node/and-001.ttl", false)

	// Open the file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Parse the Turtle data
	err = g.Parse(f, "text/turtle")
	if err != nil {
		return nil, err
	}

	return g, nil
}

func TestRDFUnmarshalling(t *testing.T) {
	// Load the test graph
	g, err := loadGraph("../data-shapes/data-shapes-test-suite/tests/core/node/and-001.ttl")
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Create an unmarshaller
	u := rdft.NewUnmarshaller(g)

	// Example 1: Unmarshal a valid rectangle
	var validRect Rectangle
	err = u.Unmarshal("http://datashapes.org/sh/tests/core/node/and-001.test#ValidRectangle1", &validRect)
	if err != nil {
		t.Fatalf("Failed to unmarshal valid rectangle: %v", err)
	}

	fmt.Printf("Valid Rectangle: %+v\n", validRect)
	if validRect.Width != 2 || validRect.Height != 3 {
		t.Errorf("Expected width=2, height=3, got width=%d, height=%d", validRect.Width, validRect.Height)
	}

	// Example 2: Unmarshal an invalid rectangle (missing width)
	var invalidRect1 Rectangle
	err = u.Unmarshal("http://datashapes.org/sh/tests/core/node/and-001.test#InvalidRectangle1", &invalidRect1)
	if err != nil {
		t.Fatalf("Failed to unmarshal invalid rectangle 1: %v", err)
	}

	fmt.Printf("Invalid Rectangle 1: %+v\n", invalidRect1)
	if invalidRect1.Width != 0 || invalidRect1.Height != 3 {
		t.Errorf("Expected width=0, height=3, got width=%d, height=%d", invalidRect1.Width, invalidRect1.Height)
	}

	// Example 3: Unmarshal an invalid rectangle (missing height)
	var invalidRect2 Rectangle
	err = u.Unmarshal("http://datashapes.org/sh/tests/core/node/and-001.test#InvalidRectangle2", &invalidRect2)
	if err != nil {
		t.Fatalf("Failed to unmarshal invalid rectangle 2: %v", err)
	}

	fmt.Printf("Invalid Rectangle 2: %+v\n", invalidRect2)
	if invalidRect2.Width != 2 || invalidRect2.Height != 0 {
		t.Errorf("Expected width=2, height=0, got width=%d, height=%d", invalidRect2.Width, invalidRect2.Height)
	}

	// Example 4: Unmarshal the test case
	// From our debugging, we found the correct URI is:
	testCaseURI := "http://datashapes.org/sh/tests/core/node/and-001"
	var testCase TestCase
	err = u.Unmarshal(testCaseURI, &testCase)
	if err != nil {
		t.Fatalf("Failed to unmarshal test case: %v", err)
	}

	fmt.Printf("Test Case: %+v\n", testCase)
	if testCase.Label != "Test of sh:and at node shape 001" {
		t.Errorf("Expected label 'Test of sh:and at node shape 001', got '%s'", testCase.Label)
	}

	// Example 5: Unmarshal the manifest
	// The manifest is the base URI itself
	manifestURI := "http://datashapes.org/sh/tests/core/node/and-001.ttl"
	var manifest Manifest
	err = u.Unmarshal(manifestURI, &manifest)
	if err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	fmt.Printf("Manifest: %+v\n", manifest)
	if len(manifest.Entries) == 0 {
		t.Errorf("Expected manifest to have entries")
	}

	// Example 6: Get all values for a predicate
	values := u.GetValues(testCaseURI, "http://www.w3.org/2000/01/rdf-schema#label")
	fmt.Printf("Label values: %v\n", values)
	if len(values) == 0 {
		t.Errorf("Expected to find label values")
	}

	// Example 7: Get a string value
	label, err := u.GetString(testCaseURI, "http://www.w3.org/2000/01/rdf-schema#label")
	if err != nil {
		t.Fatalf("Failed to get label: %v", err)
	}
	fmt.Printf("Label: %s\n", label)
	if label != "Test of sh:and at node shape 001" {
		t.Errorf("Expected label 'Test of sh:and at node shape 001', got '%s'", label)
	}
}

// Example of a struct with localized text
type LocalizedExample struct {
	URI         string               `rdf:"@id"`
	Title       rdft.LocalizedString `rdf:"http://example.org/title" rdf-type:"localized"`
	Description *rdft.LocalizedText  `rdf:"http://example.org/description" rdf-type:"localized"`
}

// Example of a custom unmarshaller
type CustomResource struct {
	URI        string
	Properties map[string][]rdft.Value
}

// UnmarshalRDF implements the RDFUnmarshaler interface
func (cr *CustomResource) UnmarshalRDF(values []rdft.Value) error {
	cr.Properties = make(map[string][]rdft.Value)

	// Process all values
	for _, value := range values {
		// Extract predicate from the value (this is just an example)
		predicate := "unknown"
		if res, ok := value.(*rdft.Resource); ok {
			parts := strings.Split(res.URI, "#")
			if len(parts) > 1 {
				predicate = parts[1]
			}
		}

		// Add to properties
		cr.Properties[predicate] = append(cr.Properties[predicate], value)
	}

	return nil
}

// TestResourceToStringFailure tests that unmarshalling a Resource into a string field fails
func TestResourceToStringFailure(t *testing.T) {
	// Load the test graph
	g, err := loadGraph("../data-shapes/data-shapes-test-suite/tests/core/node/and-001.ttl")
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Create an unmarshaller
	u := rdft.NewUnmarshaller(g)

	// Define a struct with a string field for a resource
	type InvalidStruct struct {
		URI  string `rdf:"@id"`
		Type string `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"` // This should be rdft.Resource
	}

	// Try to unmarshal a resource into a string field
	var invalid InvalidStruct
	err = u.Unmarshal("http://datashapes.org/sh/tests/core/node/and-001", &invalid)

	// Verify that an error occurred
	if err == nil {
		t.Errorf("Expected error when unmarshalling resource into string field, but got nil")
	} else {
		fmt.Printf("Correctly failed with error: %v\n", err)
	}
}

// TestStrictMode tests that strict mode returns an error for unknown predicates
func TestStrictMode(t *testing.T) {
	// Load the test graph
	g, err := loadGraph("../data-shapes/data-shapes-test-suite/tests/core/node/and-001.ttl")
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Define a struct with only some of the fields from the test case
	type PartialTestCase struct {
		URI   rdft.Resource `rdf:"@id"`
		Label string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		// Deliberately missing other fields like Type, Action, Result, Status
	}

	// Test 1: Without strict mode - should succeed
	u := rdft.NewUnmarshaller(g)
	var partial1 PartialTestCase
	err = u.Unmarshal("http://datashapes.org/sh/tests/core/node/and-001", &partial1)
	if err != nil {
		t.Errorf("Expected success without strict mode, but got error: %v", err)
	} else {
		fmt.Printf("Successfully unmarshalled without strict mode: %+v\n", partial1)
	}

	// Test 2: With strict mode - should fail with unknown predicates
	uStrict := rdft.NewUnmarshaller(g, rdft.WithStrictMode())
	var partial2 PartialTestCase
	err = uStrict.Unmarshal("http://datashapes.org/sh/tests/core/node/and-001", &partial2)
	if err == nil {
		t.Errorf("Expected error with strict mode, but got nil")
	} else {
		fmt.Printf("Correctly failed with strict mode: %v\n", err)
	}
}
