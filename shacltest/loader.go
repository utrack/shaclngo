package shacltest

import (
	"github.com/deiu/rdf2go"
	"github.com/utrack/caisson-go/errors"
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

// GetTestManifests retrieves all SHACL validation tests from the given graph
func GetTestManifests(g *rdf2go.Graph) (*Manifest, error) {
	// Create an unmarshaller for the graph
	unmarshaller := rdft.NewUnmarshaller(g)

	// Find all tests of type Validate
	testTriples := g.All(
		nil,
		rdf2go.NewResource(RDF+"type"),
		rdf2go.NewResource(MANIFEST+"Manifest"),
	)

	ret := &Manifest{}

	if len(testTriples) == 0 {
		return nil, errors.Errorf("expected 1 manifest, found 0")
	}

	if len(testTriples) > 1 {
		return nil, errors.Errorf("expected 1 manifest, found %d", len(testTriples))
	}

	err := unmarshaller.Unmarshal(testTriples[0].Subject.RawValue(), ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
