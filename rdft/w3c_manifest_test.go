package rdft

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deiu/rdf2go"
)

// TestManifestStruct represents the W3C test manifest structure
type TestManifestStruct struct {
	URI     string      `rdf:"@id"`
	Comment string      `rdf:"http://www.w3.org/2000/01/rdf-schema#comment"`
	Entries []Resource `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#entries"`
}

// TestEntryStruct represents a single test entry in the manifest
type TestEntryStruct struct {
	URI      string    `rdf:"@id"`
	Type     Resource  `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
	Name     string    `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#name"`
	Comment  string    `rdf:"http://www.w3.org/2000/01/rdf-schema#comment"`
	Action   Resource  `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#action"`
	Result   Resource  `rdf:"http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#result"`
	Approval Resource  `rdf:"http://www.w3.org/ns/rdftest#approval"`
}

// TestW3CManifestUnmarshalling tests if the rdft unmarshaller can handle the W3C Turtle test manifest
func TestW3CManifestUnmarshalling(t *testing.T) {
	// Path to the manifest file
	manifestPath := filepath.Join("testfiles", "turtle-tests-w3c-mirror", "manifest.ttl")

	// Check if the file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatalf("Manifest file not found: %s", manifestPath)
	}

	// Create a new graph
	graph := rdf2go.NewGraph("")

	// Open the file
	file, err := os.Open(manifestPath)
	if err != nil {
		t.Fatalf("Failed to open manifest file: %v", err)
	}
	defer file.Close()

	// Parse the file
	err = graph.Parse(file, "text/turtle")
	if err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	// Print the number of triples
	count := graph.Len()
	t.Logf("Parsed manifest with %d triples", count)

	// If there are no triples, the parsing failed
	if count == 0 {
		t.Fatalf("No triples found in the manifest file")
	}

	// Try to find the manifest node
	manifestNode := rdf2go.NewResource("")
	entriesNode := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#entries")

	u := NewUnmarshaller(graph)

	// Find all subjects that have the entries predicate
	entriesTriples := graph.All(nil, entriesNode, nil)
	if len(entriesTriples) == 0 {
		t.Fatalf("No entries found in the manifest")
	}

	// Use the first subject as the manifest node
	manifestNode = entriesTriples[0].Subject
	t.Logf("Found manifest node: %s", manifestNode)

	// Try to unmarshal the manifest
	manifest := &TestManifestStruct{}
	err = u.Unmarshal(manifestNode.RawValue(), manifest)
	if err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	// Check if the manifest has entries
	t.Logf("Manifest comment: %s", manifest.Comment)
	t.Logf("Manifest has %d entries", len(manifest.Entries))

	// If there are no entries, the unmarshalling failed
	if len(manifest.Entries) == 0 {
		t.Fatalf("No entries found in the manifest after unmarshalling")
	}

	// Try to unmarshal the first few test entries
	for i := range manifest.Entries {
		testResource := manifest.Entries[i]

		// Try to unmarshal the test entry
		testEntry := &TestEntryStruct{}
		err = u.Unmarshal(testResource.URI, testEntry)
		if err != nil {
			t.Fatalf("Failed to unmarshal test entry %s: %v", testResource.URI, err)
		}

		// Print the test details
		t.Logf("Test %d: ID=%s, Name=%s, Type=%s", i+1, testEntry.URI, testEntry.Name, testEntry.Type.URI)
		t.Logf("  Action: %s", testEntry.Action.URI)
		t.Logf("  Result: %s", testEntry.Result.URI)
		t.Logf("  Approval: %s", testEntry.Approval.URI)
	}

	// Try to manually extract test details for the first test
	if len(manifest.Entries) > 0 {
		testResource := manifest.Entries[0]
		testNode := rdf2go.NewResource(testResource.URI)

		// Get all triples for this test
		testTriples := graph.All(testNode, nil, nil)
		t.Logf("Found %d triples for test %s", len(testTriples), testResource.URI)

		// Print the first few triples
		for i := range testTriples {
			triple := testTriples[i]
			t.Logf("  Triple %d: %s %s %s", i+1, triple.Predicate, getObjectType(triple.Object), triple.Object)
		}
	}
}

// getObjectType returns the type of an RDF object
func getObjectType(obj rdf2go.Term) string {
	switch obj.(type) {
	case *rdf2go.Resource:
		return "Resource"
	case *rdf2go.Literal:
		return "Literal"
	case *rdf2go.BlankNode:
		return "BlankNode"
	default:
		return "Unknown"
	}
}
