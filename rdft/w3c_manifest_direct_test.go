package rdft

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deiu/rdf2go"
)

// TestW3CManifestDirectParsing tests direct parsing of the W3C Turtle test manifest
// without relying on the rdft unmarshaller
func TestW3CManifestDirectParsing(t *testing.T) {
	// Path to the manifest file
	manifestPath := filepath.Join("testfiles", "turtle-tests-w3c-mirror", "manifest.ttl")
	manifestDir := filepath.Dir(manifestPath)
	manifestURI := "file://" + manifestPath

	// Check if the file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatalf("Manifest file not found: %s", manifestPath)
	}

	// Create a new graph with the manifest URI as base
	graph := rdf2go.NewGraph(manifestURI)

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

	// Try to find the manifest node by its type (mf:Manifest)
	typePred := rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
	manifestType := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#Manifest")
	entriesNode := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#entries")

	// Find all nodes of type mf:Manifest
	manifestTriples := graph.All(nil, typePred, manifestType)
	if len(manifestTriples) == 0 {
		t.Fatalf("No manifest node found in the graph")
	}

	// Use the first manifest node
	manifestNode := manifestTriples[0].Subject
	t.Logf("Found manifest node: %s", manifestNode)

	// Find the entries triple for the manifest node
	entriesTriples := graph.All(manifestNode, entriesNode, nil)
	if len(entriesTriples) == 0 {
		t.Fatalf("No entries found for the manifest node")
	}

	// Get the entries object
	entriesObj := entriesTriples[0].Object
	t.Logf("Found entries node: %s", entriesObj)

	// The entries object should be a blank node for a collection
	var entriesList []rdf2go.Term

	switch node := entriesObj.(type) {
	case *rdf2go.BlankNode:
		// This is a blank node, which might be the start of an RDF collection
		// We need to traverse the collection to get all the entries
		entriesList = traverseRDFCollectionForTest(graph, node)
	default:
		t.Fatalf("Unexpected node type for entries: %T", entriesObj)
	}

	t.Logf("Found %d entries in the manifest", len(entriesList))

	// If there are no entries, the parsing failed
	if len(entriesList) == 0 {
		t.Fatalf("No entries found in the manifest after parsing")
	}

	// Try to extract details for the first few test entries
	for i := 0; i < 5 && i < len(entriesList); i++ {
		testNode := entriesList[i]

		// Get all triples for this test
		var testTriples []*rdf2go.Triple

		switch node := testNode.(type) {
		case *rdf2go.Resource:
			testTriples = graph.All(node, nil, nil)
		case *rdf2go.BlankNode:
			testTriples = graph.All(node, nil, nil)
		default:
			t.Fatalf("Unexpected node type for test: %T", testNode)
		}

		t.Logf("Found %d triples for test %s", len(testTriples), testNode)

		// Extract test details
		testDetails := extractTestDetailsForTest(graph, testNode)

		// Print the test details
		t.Logf("Test %d: ID=%s, Name=%s, Type=%s", i+1, testDetails["id"], testDetails["name"], testDetails["type"])
		t.Logf("  Action: %s", testDetails["action"])
		t.Logf("  Result: %s", testDetails["result"])

		// Check if the test files exist
		actionFile := testDetails["action"]
		if actionFile != "" {
			// Convert URI to file path
			actionFile = strings.TrimPrefix(actionFile, "file://")
			// Extract the filename from the path to avoid duplication
			_, actionFileName := filepath.Split(actionFile)
			actionFile = filepath.Join(manifestDir, actionFileName)

			if _, err := os.Stat(actionFile); os.IsNotExist(err) {
				t.Logf("  Warning: Action file not found: %s", actionFile)
			} else {
				t.Logf("  Action file exists: %s", actionFile)
			}
		}

		resultFile := testDetails["result"]
		if resultFile != "" {
			// Convert URI to file path
			resultFile = strings.TrimPrefix(resultFile, "file://")
			// Extract the filename from the path to avoid duplication
			_, resultFileName := filepath.Split(resultFile)
			resultFile = filepath.Join(manifestDir, resultFileName)

			if _, err := os.Stat(resultFile); os.IsNotExist(err) {
				t.Logf("  Warning: Result file not found: %s", resultFile)
			} else {
				t.Logf("  Result file exists: %s", resultFile)
			}
		}
	}
}

// traverseRDFCollectionForTest traverses an RDF collection and returns all its elements
func traverseRDFCollectionForTest(graph *rdf2go.Graph, startNode rdf2go.Term) []rdf2go.Term {
	var result []rdf2go.Term

	// RDF collection predicates
	firstPred := rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#first")
	restPred := rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest")
	nilNode := rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#nil")

	currentNode := startNode
	for {
		// Get the first element
		firstTriples := graph.All(currentNode, firstPred, nil)
		if len(firstTriples) == 0 {
			break
		}

		// Add the element to the result
		result = append(result, firstTriples[0].Object)

		// Get the rest of the collection
		restTriples := graph.All(currentNode, restPred, nil)
		if len(restTriples) == 0 {
			break
		}

		// Check if we've reached the end of the collection
		if restTriples[0].Object.Equal(nilNode) {
			break
		}

		// Move to the next node
		currentNode = restTriples[0].Object
	}

	return result
}

// extractTestDetailsForTest extracts details for a test node
func extractTestDetailsForTest(graph *rdf2go.Graph, testNode rdf2go.Term) map[string]string {
	details := make(map[string]string)

	// Common predicates
	typePred := rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
	namePred := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#name")
	commentPred := rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#comment")
	actionPred := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#action")
	resultPred := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#result")
	approvalPred := rdf2go.NewResource("http://www.w3.org/ns/rdftest#approval")

	// Set the ID
	switch node := testNode.(type) {
	case *rdf2go.Resource:
		details["id"] = node.URI
	case *rdf2go.BlankNode:
		details["id"] = "_:" + node.ID
	}

	// Get the type
	typeTriples := graph.All(testNode, typePred, nil)
	if len(typeTriples) > 0 {
		if res, ok := typeTriples[0].Object.(*rdf2go.Resource); ok {
			details["type"] = res.URI
		}
	}

	// Get the name
	nameTriples := graph.All(testNode, namePred, nil)
	if len(nameTriples) > 0 {
		if lit, ok := nameTriples[0].Object.(*rdf2go.Literal); ok {
			details["name"] = lit.Value
		}
	}

	// Get the comment
	commentTriples := graph.All(testNode, commentPred, nil)
	if len(commentTriples) > 0 {
		if lit, ok := commentTriples[0].Object.(*rdf2go.Literal); ok {
			details["comment"] = lit.Value
		}
	}

	// Get the action
	actionTriples := graph.All(testNode, actionPred, nil)
	if len(actionTriples) > 0 {
		if res, ok := actionTriples[0].Object.(*rdf2go.Resource); ok {
			details["action"] = res.URI
		}
	}

	// Get the result
	resultTriples := graph.All(testNode, resultPred, nil)
	if len(resultTriples) > 0 {
		if res, ok := resultTriples[0].Object.(*rdf2go.Resource); ok {
			details["result"] = res.URI
		}
	}

	// Get the approval
	approvalTriples := graph.All(testNode, approvalPred, nil)
	if len(approvalTriples) > 0 {
		if res, ok := approvalTriples[0].Object.(*rdf2go.Resource); ok {
			details["approval"] = res.URI
		}
	}

	return details
}
