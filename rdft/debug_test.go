package rdft_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/deiu/rdf2go"
)

// loadGraph loads an RDF graph from a Turtle file
func loadDebugGraph(path string) (*rdf2go.Graph, error) {
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

	// Print the graph size directly after parsing
	fmt.Printf("Graph size after parsing: %d triples\n", len(g.IterTriples()))

	return g, nil
}

func TestDebugRDFGraph(t *testing.T) {
	// Load the test graph
	g, err := loadDebugGraph("../data-shapes/data-shapes-test-suite/tests/core/node/and-001.ttl")
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Print graph size using different methods
	fmt.Printf("Graph size using All(): %d triples\n", len(g.All(nil, nil, nil)))
	fmt.Printf("Graph size using IterTriples(): %d triples\n", len(g.IterTriples()))

	// Print all subjects
	fmt.Println("\n=== All subjects in the graph ===")
	subjects := make(map[string]bool)
	for _, triple := range g.All(nil, nil, nil) {
		subjects[triple.Subject.String()] = true
	}
	for subject := range subjects {
		fmt.Println(subject)
	}

	// Print all triples with rdfs:label
	fmt.Println("\n=== All triples with rdfs:label ===")
	labelTriples := g.All(
		nil,
		rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#label"),
		nil,
	)
	for _, triple := range labelTriples {
		fmt.Printf("Subject: %s\nObject: %s\n\n", triple.Subject, triple.Object)
	}

	// Try different ways to access the test case
	fmt.Println("\n=== Trying different ways to access the test case ===")
	
	// Try with different URI formats
	possibleURIs := []string{
		"and-001",
		"http://datashapes.org/sh/tests/core/node/and-001",
		"http://datashapes.org/sh/tests/core/node/and-001.ttl#and-001",
		"http://datashapes.org/sh/tests/core/node/and-001/and-001",
	}
	
	for _, uri := range possibleURIs {
		resource := rdf2go.NewResource(uri)
		triples := g.All(resource, nil, nil)
		fmt.Printf("URI: %s - Found %d triples\n", uri, len(triples))
		
		for _, triple := range triples {
			fmt.Printf("  %s %s %s\n", triple.Subject, triple.Predicate, triple.Object)
		}
	}
	
	// Try to find triples with rdfs:label directly
	fmt.Println("\n=== Direct access to rdfs:label triples ===")
	rdfsLabel := rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#label")
	// Use a different variable name to avoid shadowing the previous labelTriples
	directLabelTriples := g.All(nil, rdfsLabel, nil)
	fmt.Printf("Found %d rdfs:label triples\n", len(directLabelTriples))
	
	for _, triple := range directLabelTriples {
		fmt.Printf("  %s %s %s\n", triple.Subject, triple.Predicate, triple.Object)
	}
}
