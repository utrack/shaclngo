package shaclngo

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	rdf "github.com/deiu/gon3"
	"github.com/deiu/rdf2go"
)

func TestManual(t *testing.T) {
	ownPrefix := "http://example.com/test-suite/"
	dir := "data-shapes/data-shapes-test-suite/tests/core/"
	g := rdf2go.NewGraph(ownPrefix, false)
	f, err := os.Open(dir + "manifest.ttl")
	if err != nil {
		t.Fatal(err)
	}
	err = g.Parse(f, "text/turtle")
	if err != nil {
		t.Fatal(err)
	}

	loadedIncludes := make(map[string]bool)
	for {
		includes := g.All(nil, rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#include"), nil)
		hadNewIncludes := false
		for _, include := range includes {
			subj := include.Subject.RawValue()
			asStr := include.Object.RawValue()
			if loadedIncludes[asStr] {
				continue
			}
			fmt.Println("object ", asStr, " for subject ", subj)
			if strings.HasPrefix(asStr, ownPrefix) {
				fPath := strings.TrimPrefix(asStr, ownPrefix)
				f, err := os.Open(dir + fPath)
				if err != nil {
					t.Fatal(err)
				}
				newNs := ownPrefix + path.Dir(fPath)
				if newNs[len(newNs)-1] != '/' {
					newNs += "/"
				}

				parser, err := rdf.NewParser(ownPrefix).Parse(f)
				if err != nil {
					t.Fatal(err)
				}
				for s := range parser.IterTriples() {
					if strings.HasPrefix(s.Object.RawValue(), ownPrefix) {
						s.Object = rdf.NewIRI(newNs + strings.TrimPrefix(s.Object.RawValue(), ownPrefix))
					}

					g.AddTriple(rdf2go.NewResource(s.Subject.RawValue()), rdf2go.NewResource(s.Predicate.RawValue()), rdf2go.NewResource(s.Object.RawValue()))
				}
			} else {
				fmt.Println("URI load ", asStr)
				err := g.LoadURI(asStr)
				if err != nil {
					t.Fatal(err)
				}
			}
			fmt.Println("loaded ", asStr)
			loadedIncludes[asStr] = true
			hadNewIncludes = true
		}
		if !hadNewIncludes {
			break
		}
	}

	var buf bytes.Buffer
	err = g.Serialize(&buf, "text/turtle")
	if err != nil {
		t.Fatal(err)
	}

}
