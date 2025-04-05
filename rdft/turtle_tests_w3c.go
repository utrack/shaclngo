package rdft

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/deiu/rdf2go"
)

// W3CTurtleTest represents a single W3C Turtle test case
type W3CTurtleTest struct {
	ID         string // The test ID (e.g., "#IRI_subject")
	Name       string // The test name
	Comment    string // The test description
	Type       string // The test type (e.g., "TestTurtleEval")
	Approval   string // The approval status
	ActionFile string // The Turtle file to test
	ResultFile string // The expected output file (usually NT format)
}

// W3CTurtleTestManifest represents the collection of W3C Turtle tests
type W3CTurtleTestManifest struct {
	BaseDir string           // Base directory for test files
	Tests   []*W3CTurtleTest // List of tests
}

// LoadW3CTurtleTestManifest loads the W3C Turtle test manifest from the specified directory
func LoadW3CTurtleTestManifest(manifestPath string) (*W3CTurtleTestManifest, error) {
	// Create a new manifest
	manifest := &W3CTurtleTestManifest{
		BaseDir: filepath.Dir(manifestPath),
		Tests:   make([]*W3CTurtleTest, 0),
	}

	fmt.Printf("Loading manifest from %s\n", manifestPath)

	// Create a new graph with the manifest URI as base
	manifestURI := "file://" + manifestPath
	graph := rdf2go.NewGraph(manifestURI)

	// Open the file
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %v", err)
	}
	defer file.Close()

	// Parse the file
	err = graph.Parse(file, "text/turtle")
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %v", err)
	}

	// Try to find the manifest node by its type (mf:Manifest)
	typePred := rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
	manifestType := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#Manifest")
	entriesNode := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#entries")

	// Find all nodes of type mf:Manifest
	manifestTriples := graph.All(nil, typePred, manifestType)
	if len(manifestTriples) == 0 {
		return nil, fmt.Errorf("no manifest node found in the graph")
	}

	// Use the first manifest node
	manifestNode := manifestTriples[0].Subject
	fmt.Printf("Found manifest node: %s\n", manifestNode)

	// Find the entries triple for the manifest node
	entriesTriples := graph.All(manifestNode, entriesNode, nil)
	if len(entriesTriples) == 0 {
		return nil, fmt.Errorf("no entries found for the manifest node")
	}

	// Get the entries object
	entriesObj := entriesTriples[0].Object
	fmt.Printf("Found entries node: %s\n", entriesObj)

	// The entries object should be a blank node for a collection
	var entriesList []rdf2go.Term

	switch node := entriesObj.(type) {
	case *rdf2go.BlankNode:
		// This is a blank node, which might be the start of an RDF collection
		// We need to traverse the collection to get all the entries
		entriesList = traverseRDFCollection(graph, node)
	default:
		return nil, fmt.Errorf("unexpected node type for entries: %T", entriesObj)
	}

	fmt.Printf("Found %d entries in the manifest\n", len(entriesList))

	// If there are no entries, the parsing failed
	if len(entriesList) == 0 {
		return nil, fmt.Errorf("no entries found in the manifest after parsing")
	}

	// Process each test entry
	for i, testNode := range entriesList {
		// Extract test details
		test, err := extractTestDetailsFromNode(graph, testNode)
		if err != nil {
			return nil, fmt.Errorf("failed to extract test details for entry %d: %v", i+1, err)
		}

		// Fix file paths
		if test.ActionFile != "" {
			// Convert URI to file path
			test.ActionFile = strings.TrimPrefix(test.ActionFile, "file://")
			// Extract the filename from the path to avoid duplication
			_, actionFileName := filepath.Split(test.ActionFile)
			test.ActionFile = actionFileName
		}

		if test.ResultFile != "" {
			// Convert URI to file path
			test.ResultFile = strings.TrimPrefix(test.ResultFile, "file://")
			// Extract the filename from the path to avoid duplication
			_, resultFileName := filepath.Split(test.ResultFile)
			test.ResultFile = resultFileName
		}

		manifest.Tests = append(manifest.Tests, test)
	}

	return manifest, nil
}

// traverseRDFCollection traverses an RDF collection and returns all its elements
func traverseRDFCollection(graph *rdf2go.Graph, startNode rdf2go.Term) []rdf2go.Term {
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

// extractTestDetailsFromNode extracts the details of a test from a node in the graph
func extractTestDetailsFromNode(graph *rdf2go.Graph, testNode rdf2go.Term) (*W3CTurtleTest, error) {
	test := &W3CTurtleTest{}

	// Set the ID
	switch node := testNode.(type) {
	case *rdf2go.Resource:
		test.ID = node.URI
	case *rdf2go.BlankNode:
		test.ID = "_:" + node.ID
	default:
		return nil, fmt.Errorf("unexpected node type for test: %T", testNode)
	}

	// Common predicates
	typePred := rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
	namePred := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#name")
	commentPred := rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#comment")
	actionPred := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#action")
	resultPred := rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#result")
	approvalPred := rdf2go.NewResource("http://www.w3.org/ns/rdftest#approval")

	// Get the type
	typeTriples := graph.All(testNode, typePred, nil)
	if len(typeTriples) > 0 {
		if res, ok := typeTriples[0].Object.(*rdf2go.Resource); ok {
			test.Type = res.URI
			// Extract the type name from the URI
			parts := strings.Split(res.URI, "#")
			if len(parts) > 1 {
				test.Type = parts[1]
			}
		}
	}

	// Get the name
	nameTriples := graph.All(testNode, namePred, nil)
	if len(nameTriples) > 0 {
		if lit, ok := nameTriples[0].Object.(*rdf2go.Literal); ok {
			test.Name = lit.Value
		}
	}

	// Get the comment
	commentTriples := graph.All(testNode, commentPred, nil)
	if len(commentTriples) > 0 {
		if lit, ok := commentTriples[0].Object.(*rdf2go.Literal); ok {
			test.Comment = lit.Value
		}
	}

	// Get the action
	actionTriples := graph.All(testNode, actionPred, nil)
	if len(actionTriples) > 0 {
		if res, ok := actionTriples[0].Object.(*rdf2go.Resource); ok {
			test.ActionFile = res.URI
		}
	}

	// Get the result
	resultTriples := graph.All(testNode, resultPred, nil)
	if len(resultTriples) > 0 {
		if res, ok := resultTriples[0].Object.(*rdf2go.Resource); ok {
			test.ResultFile = res.URI
		}
	}

	// Get the approval
	approvalTriples := graph.All(testNode, approvalPred, nil)
	if len(approvalTriples) > 0 {
		if res, ok := approvalTriples[0].Object.(*rdf2go.Resource); ok {
			test.Approval = res.URI
			// Extract the approval name from the URI
			parts := strings.Split(res.URI, "#")
			if len(parts) > 1 {
				test.Approval = parts[1]
			}
		}
	}

	return test, nil
}

// RunTest runs a single W3C Turtle test
func (test *W3CTurtleTest) RunTest(baseDir string) (bool, error) {
	// Check if the files exist
	actionFilePath := filepath.Join(baseDir, test.ActionFile)
	resultFilePath := filepath.Join(baseDir, test.ResultFile)

	// Check if the action file exists
	if _, err := os.Stat(actionFilePath); os.IsNotExist(err) {
		return false, fmt.Errorf("action file not found: %s", actionFilePath)
	}

	// result file can be empty, but only for PositiveSyntax and NegativeSyntax tests
	resultFileDescribed := true
	if test.Type == "TestTurtlePositiveSyntax" || test.Type == "TestTurtleNegativeSyntax" {
		if test.ResultFile == "" {
			resultFileDescribed = false
		}
	}

	if resultFileDescribed {
		if _, err := os.Stat(resultFilePath); os.IsNotExist(err) {
			// If the test type requires a result file but it doesn't exist, report an error
			return false, fmt.Errorf("result file not found: %s", resultFilePath)
		}
	}

	// For negative syntax tests, we expect parsing to fail
	if test.Type == "TestTurtleNegativeSyntax" {
		// Try to parse the action file

		actionFile, err := os.Open(actionFilePath)
		if err != nil {
			return false, fmt.Errorf("failed to open action file: %v", err)
		}
		defer actionFile.Close()

		// If parsing succeeds for a negative syntax test, the test fails
		actionGraph := rdf2go.NewGraph("")
		err = actionGraph.Parse(actionFile, "text/turtle")
		if err != nil {
			// Parsing failed as expected for a negative test
			return true, nil
		} else {
			// Parsing succeeded, but should have failed
			return false, nil
		}
	}

	// For positive tests, we parse both files and compare the results
	// Parse the action file
	actionGraph := rdf2go.NewGraph("")

	// Open the action file
	actionFile, err := os.Open(actionFilePath)
	if err != nil {
		return false, fmt.Errorf("failed to open action file: %v", err)
	}
	defer actionFile.Close()

	// Parse the file
	err = actionGraph.Parse(actionFile, "text/turtle")
	if err != nil {
		return false, fmt.Errorf("failed to parse action file: %v", err)
	}

	if !resultFileDescribed {
		return true, nil
	}

	// Parse the result file
	resultGraph := rdf2go.NewGraph("")

	// Open the result file
	resultFile, err := os.Open(resultFilePath)
	if err != nil {
		return false, fmt.Errorf("failed to open result file: %v", err)
	}
	defer resultFile.Close()

	// Parse the file - result files are in N-Triples format
	// Try different MIME types for N-Triples
	err = resultGraph.Parse(resultFile, "application/n-triples")
	if err != nil {
		// Try alternate MIME type
		resultFile.Seek(0, 0) // Reset file position
		err = resultGraph.Parse(resultFile, "text/plain")
	}
	if err != nil {
		// Try another alternate MIME type
		resultFile.Seek(0, 0) // Reset file position
		err = resultGraph.Parse(resultFile, "text/turtle")
	}
	if err != nil {
		return false, fmt.Errorf("failed to parse result file: %v", err)
	}

	// Compare the graphs
	return compareStructGraphs(actionGraph, resultGraph)
}

func compareStructGraphs(g1, g2 *rdf2go.Graph) (bool, error) {
	var got1, got2 map[string]any

	err := NewUnmarshaller(g1).Unmarshal("", &got1)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal action graph: %v", err)
	}

	err = NewUnmarshaller(g2).Unmarshal("", &got2)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal result graph: %v", err)
	}

	return reflect.DeepEqual(got1, got2), nil
}
