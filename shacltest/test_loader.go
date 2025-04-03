package shacltest

import (
	"fmt"
	"strings"

	"github.com/deiu/rdf2go"
	"github.com/utrack/shaclngo/bitgraph"
)

type Test struct {
	ID          string
	Description string
	Status      string

	Action         Action
	ExpectedResult ResultManifest
}

type Action struct {
	Data   rdf2go.Term
	Shapes rdf2go.Term
}

type ResultManifest struct {
	Conforms bool
	Results  []Result
}

type Result struct {
	FocusNode                 rdf2go.Term
	ResultSeverity            rdf2go.Term
	SourceConstraintComponent rdf2go.Term
	SourceShape               rdf2go.Term
	Value                     rdf2go.Term
}

func getTestManifests(g *bitgraph.Graph) []Test {
	tests := g.Filter(nil, rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource("http://www.w3.org/ns/shacl-test#Validate"))

	var ret []Test
	for _, testSubject := range tests {

		t := Test{
			ID: testSubject.Subject.RawValue(),
		}

		tsEntries := g.Filter(testSubject.Subject, nil, nil)

		for _, entry := range tsEntries {
			switch entry.Predicate.RawValue() {
			case "http://www.w3.org/2000/01/rdf-schema#label":
				t.Description = entry.Object.RawValue()
			case "http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#status":
				t.Status = strings.TrimPrefix(entry.Object.RawValue(), "http://www.w3.org/ns/shacl-test#")
			case "http://www.w3.org/1999/02/22-rdf-syntax-ns#type":
				// type is already filtered, Validate
			case "http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#action":
				actionSubj := entry.Object
				actionEntries := g.Filter(actionSubj, nil, nil)
				fmt.Println(actionSubj)
				for _, actionEntry := range actionEntries {
					switch actionEntry.Predicate.RawValue() {
					case "http://www.w3.org/ns/shacl-test#shapesGraph":
						t.Action.Shapes = actionEntry.Object
					case "http://www.w3.org/ns/shacl-test#dataGraph":
						t.Action.Data = actionEntry.Object
					default:
						fmt.Println("unknown action predicate:", actionEntry.Subject.RawValue(), " -> ", actionEntry.Predicate.RawValue())
					}
				}
			case "http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#result":
				rManifestSubj := entry.Object
				rManifestEntries := g.Filter(rManifestSubj, nil, nil)
				var manifest ResultManifest
				for _, resultEntry := range rManifestEntries {
					switch resultEntry.Predicate.RawValue() {
					case "http://www.w3.org/1999/02/22-rdf-syntax-ns#type":
						if resultEntry.Object.RawValue() != "http://www.w3.org/ns/shacl#ValidationReport" {
							fmt.Println("unexpected ValidationReport type:", resultEntry.Subject.RawValue(), " -> ", resultEntry.Object.RawValue())
						}
					case "http://www.w3.org/ns/shacl#conforms":
						manifest.Conforms = resultEntry.Object.RawValue() == "true"
					case "http://www.w3.org/ns/shacl#result":
						resultSubj := resultEntry.Object
						resultEntries := g.Filter(resultSubj, nil, nil)
						var result Result
						for _, resultEntry := range resultEntries {
							switch resultEntry.Predicate.RawValue() {
							case "http://www.w3.org/1999/02/22-rdf-syntax-ns#type":
								if resultEntry.Object.RawValue() != "http://www.w3.org/ns/shacl#ValidationResult" {
									fmt.Println("unexpected result type:", resultEntry.Subject.RawValue(), " -> ", resultEntry.Object.RawValue())
								}
							case "http://www.w3.org/ns/shacl#focusNode":
								result.FocusNode = resultEntry.Object
							case "http://www.w3.org/ns/shacl#resultSeverity":
								result.ResultSeverity = resultEntry.Object
							case "http://www.w3.org/ns/shacl#sourceConstraintComponent":
								result.SourceConstraintComponent = resultEntry.Object
							case "http://www.w3.org/ns/shacl#sourceShape":
								result.SourceShape = resultEntry.Object
							case "http://www.w3.org/ns/shacl#value":
								result.Value = resultEntry.Object
							default:
								fmt.Println("unknown result predicate:", resultEntry.Subject.RawValue(), " -> ", resultEntry.Predicate.RawValue())
							}
						}
						manifest.Results = append(manifest.Results, result)
					default:
						fmt.Println("unknown result manifest predicate:", resultEntry.Subject.RawValue(), " -> ", resultEntry.Predicate.RawValue())
					}
				}
				t.ExpectedResult = manifest
			default:
				fmt.Println("unknown predicate:", entry.Predicate.RawValue())
			}
		}

		ret = append(ret, t)
	}

	return ret
}
