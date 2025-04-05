package rdft

import (
	"github.com/deiu/rdf2go"
	"github.com/utrack/caisson-go/errors"
)

// checkNodeType verifies that a node has the expected RDF type
// Returns nil if:
// - No type assertion is required (expectedType is empty)
// - The node has the expected type
// Returns an error if the node doesn't have the expected type
func (u *Unmarshaller) checkNodeType(triples []*rdf2go.Triple, expectedType string) error {
	if expectedType == "" {
		return nil // No type assertion required
	}

	// Check if any of the types match the expected type
	for _, triple := range triples {
		if triple.Predicate.String() != "<http://www.w3.org/1999/02/22-rdf-syntax-ns#type>" {
			continue
		}

		if obj, ok := triple.Object.(*rdf2go.Resource); ok {
			if obj.URI == expectedType {
				return nil // Type matches
			}
			return errors.Errorf("node's type %s does not have expected RDF type %s", obj.URI, expectedType)
		}
	}

	// No matching type found
	return errors.Errorf("node has no RDF type, expected %s", expectedType)
}
