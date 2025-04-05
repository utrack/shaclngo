package rdft

import (
	"fmt"
	"strings"

	"github.com/deiu/rdf2go"
)

// checkNodeType verifies that a node has the expected RDF type
// Returns nil if:
// - No type assertion is required (expectedType is empty)
// - The node has the expected type
// Returns an error if the node doesn't have the expected type
func (u *Unmarshaller) checkNodeType(subject string, expectedType string) error {
	if expectedType == "" {
		return nil // No type assertion required
	}

	// Handle different subject types (resource or blank node)
	var subjectTerm rdf2go.Term
	if strings.HasPrefix(subject, "_:") {
		// This is a blank node
		blankNodeID := strings.TrimPrefix(subject, "_:")
		subjectTerm = rdf2go.NewBlankNode(blankNodeID)
	} else {
		// This is a resource
		subjectTerm = rdf2go.NewResource(subject)
	}

	// Get all rdf:type triples for this subject
	typeTriples := u.graph.All(
		subjectTerm,
		rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
		nil,
	)

	// If no type triples found, return error
	if len(typeTriples) == 0 {
		return fmt.Errorf("node %s has no RDF type, expected %s", subject, expectedType)
	}

	// Check if any of the types match the expected type
	for _, triple := range typeTriples {
		if obj, ok := triple.Object.(*rdf2go.Resource); ok {
			if obj.URI == expectedType {
				return nil // Type matches
			}
		}
	}

	// No matching type found
	return fmt.Errorf("node %s does not have expected RDF type %s (is %s instead)", subject, expectedType, typeTriples[0].Object)
}
