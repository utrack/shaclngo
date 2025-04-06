package shacltest

import (
	"fmt"
	"path"

	"github.com/deiu/rdf2go"
	"github.com/utrack/caisson-go/errors"
	"github.com/utrack/shaclngo/rdft"
	"github.com/utrack/shaclngo/rgraph"
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
func GetTestManifests(g rgraph.Graph) (*Manifest, error) {
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

	for i, et := range ret.Entries {
		fmt.Println(et.ID)
		definedByRes := rdf2go.NewResource(RDFS + "isDefinedBy")
		fileDef := g.One(rdf2go.NewResource(et.ID), definedByRes, nil)
		if fileDef == nil {
			return nil, errors.Errorf("expected 1 isDefinedBy triple for %s, found 0", et.ID)
		}

		dataRelURI := fileDef.Object.RawValue()
		shapesRelURI := fileDef.Object.RawValue()

		if et.Action.DataResource.URI != "" {
			p := path.Join(path.Dir(dataRelURI), et.Action.DataResource.URI)
			dataRelURI = p
		}
		if et.Action.ShapesResource.URI != "" {
			p := path.Join(path.Dir(shapesRelURI), et.Action.ShapesResource.URI)
			shapesRelURI = p
		}

		shapesSource := g.All(nil, definedByRes, rdf2go.NewResource(shapesRelURI))

		shapesByURI := make(map[string]struct{})

		for _, def := range shapesSource {
			types := g.All(def.Subject, rdf2go.NewResource(RDF+"type"), nil)
			isShape := false
			for _, t := range types {
				switch t.Object.RawValue() {
				case SHACL + "Shape", SHACL + "NodeShape", SHACL + "PropertyShape":
					isShape = true
				}
			}
			if isShape {
				v, _ := def.Subject.(*rdf2go.Resource)
				et.Action.ShapesObjects = append(et.Action.ShapesObjects, *v)
				shapesByURI[v.RawValue()] = struct{}{}
			}
		}

		dataSource := g.All(nil, definedByRes, rdf2go.NewResource(dataRelURI))

		for _, def := range dataSource {
			types := g.All(def.Subject, rdf2go.NewResource(RDF+"type"), nil)
			isShaped := false
			for _, t := range types {
				if _, ok := shapesByURI[t.Object.RawValue()]; ok {
					isShaped = true
				}
			}
			if isShaped {
				v, _ := def.Subject.(*rdf2go.Resource)
				et.Action.DataObjects = append(et.Action.DataObjects, *v)
			}
		}

		ret.Entries[i] = et
	}

	return ret, nil
}
