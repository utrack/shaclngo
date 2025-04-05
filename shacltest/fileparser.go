package shacltest

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/deiu/rdf2go"

	rdf "github.com/deiu/gon3"
	"github.com/utrack/caisson-go/errors"
)

func loadGraph(prefix string, path string) (*rdf2go.Graph, error) {
	g := rdf2go.NewGraph(prefix, false)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	err = g.Parse(f, "text/turtle")
	if err != nil {
		return nil, err
	}
	path = filepath.Dir(path)

	if path[len(path)-1] != '/' {
		path += "/"
	}

	loadedIncludes := make(map[string]bool)
	for {
		includes := g.All(nil, rdf2go.NewResource("http://www.w3.org/2001/sw/DataAccess/tests/test-manifest#include"), nil)
		hadNewIncludes := false
		for _, include := range includes {
			o := include.Object.RawValue()
			if loadedIncludes[o] {
				continue
			}

			if strings.HasPrefix(o, prefix) {
				pathFromPrefix := strings.TrimPrefix(o, prefix)
				rdfPrefix := prefix + pathFromPrefix
				f, err := os.Open(path + pathFromPrefix)
				if err != nil {
					return nil, err
				}

				parser, err := rdf.NewParser(rdfPrefix).Parse(f)
				if err != nil {
					return nil, err
				}
				for s := range parser.IterTriples() {
					g.AddTriple(rdf2term(s.Subject), rdf2term(s.Predicate), rdf2term(s.Object))
				}
			} else {
				err := g.LoadURI(o)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to load URI %s", o)
				}
			}
			loadedIncludes[o] = true
			hadNewIncludes = true
		}
		if !hadNewIncludes {
			break
		}
	}

	return g, nil
}

func rdf2term(term rdf.Term) rdf2go.Term {
	switch term := term.(type) {
	case *rdf.BlankNode:
		// id := fmt.Sprint(term.Id)
		return rdf2go.NewBlankNode(term.RawValue())
	case *rdf.Literal:
		if len(term.LanguageTag) > 0 {
			return rdf2go.NewLiteralWithLanguage(term.LexicalForm, term.LanguageTag)
		}
		if term.DatatypeIRI != nil && len(term.DatatypeIRI.String()) > 0 {
			return rdf2go.NewLiteralWithDatatype(term.LexicalForm, rdf2go.NewResource(debrack(term.DatatypeIRI.String())))
		}
		return rdf2go.NewLiteral(term.RawValue())
	case *rdf.IRI:
		return rdf2go.NewResource(term.RawValue())
	}
	return nil
}

// debrack removes angle brackets from a string.
func debrack(s string) string {
	if len(s) < 2 {
		return s
	}
	if s[0] != '<' {
		return s
	}
	if s[len(s)-1] != '>' {
		return s
	}
	return s[1 : len(s)-1]
}
