package rdft

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/deiu/rdf2go"
	"github.com/stretchr/testify/require"
)

// Define the validation report and result structs
type blankValidationReport struct {
	// ID is the URI of the validation report
	ID string `rdf:"@id" rdfType:"http://www.w3.org/ns/shacl-test#Validate"`

	// Results contains the validation results
	Results *ValidationResult `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#result"`
}

type ValidationResult struct {
	// ID is the URI of the validation result
	ID    string `rdf:"@id" rdfType:"http://www.w3.org/ns/shacl#ValidationReport"`
	Label string `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
}

// TestBlankNodeValidationReport tests unmarshalling a validation report from a blank node
func TestBlankNode(t *testing.T) {
	so := require.New(t)
	// Load the test TTL file
	filePath := filepath.Join("testfiles", "blank_node_test.ttl")

	g, err := loadGraph(filePath)
	so.NoError(err, "Failed to load test graph")

	// Create an unmarshaller
	unmarshaller := NewUnmarshaller(g, WithStrictMode())

	var report blankValidationReport
	err = unmarshaller.Unmarshal("file://testfiles/root-struct", &report)
	so.NoError(err, "Failed to unmarshal blank node validation report")

	so.NotNil(report.Results)
	so.Equal("A typed blank node", report.Results.Label, "Validation report label does not match")
}

// Helper function to load a graph from a file
func loadGraph(filePath string) (*rdf2go.Graph, error) {
	// Create a new graph
	g := rdf2go.NewGraph("file://" + filePath)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Parse the file
	err = g.Parse(file, "text/turtle")
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %v", err)
	}

	// Debug: Print out the number of triples in the graph
	fmt.Printf("Loaded %d triples into the graph\n", g.Len())

	for triple := range g.IterTriples() {
		fmt.Printf("Triple: %s\n", triple)
	}

	return g, nil
}
