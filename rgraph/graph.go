package rgraph

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/deiu/rdf2go"
	"github.com/utrack/caisson-go/errors"
)

type Graph interface {
	All(subject, predicate, object rdf2go.Term) []*rdf2go.Triple
	LoadURI(uri string) error
}

func NewGraph(mainPrefix string, loader Loader) Graph {
	return &rdf2goGraph{
		g: rdf2go.NewGraph(mainPrefix, false),
		l: loader,
	}
}

type rdf2goGraph struct {
	g *rdf2go.Graph
	l Loader
}

func (g *rdf2goGraph) All(subject, predicate, object rdf2go.Term) []*rdf2go.Triple {
	return g.g.All(subject, predicate, object)
}

func (g *rdf2goGraph) LoadURI(uri string) error {
	ts, err := g.l.LoadURI(uri)
	if err != nil {
		return err
	}
	for _, t := range ts {
		g.g.AddTriple(t.Subject, t.Predicate, t.Object)
	}
	return nil
}

type Loader interface {
	LoadURI(uri string) ([]*rdf2go.Triple, error)
}

type loaderFile struct {
	prefix string
	path   string
}

type loaderPfx struct {
	// TODO mutex lURIs
	loadedURIs map[string]bool
	loaders    map[string]Loader
}

func NewPrefixedLoader(loaders map[string]Loader) Loader {
	return &loaderPfx{
		loaders:    loaders,
		loadedURIs: make(map[string]bool),
	}
}

func (g *loaderPfx) LoadURI(uri string) ([]*rdf2go.Triple, error) {
	if _, ok := g.loadedURIs[uri]; ok {
		return nil, errors.Errorf("URI %s already loaded", uri)
	}

	for prefix, loader := range g.loaders {
		if strings.HasPrefix(uri, prefix) {
			ret, err := loader.LoadURI(uri)
			if err == nil {
				g.loadedURIs[uri] = true
			}
			return ret, errors.Wrap(err, "when using a loader with prefix "+prefix)
		}
	}
	return nil, errors.Errorf("failed to find loader for URI %s", uri)
}

func NewLoaderFile(prefix string, directory string) Loader {
	return &loaderFile{
		prefix: prefix,
		path:   directory,
	}
}

func (l *loaderFile) LoadURI(uri string) ([]*rdf2go.Triple, error) {
	unprefixedPath := strings.TrimPrefix(uri, l.prefix)

	pathToFile := filepath.Join(l.path, unprefixedPath)

	if i := strings.LastIndex(pathToFile, "#"); i != -1 {
		pathToFile = pathToFile[:i]
	}

	f, err := os.Open(pathToFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %s", pathToFile)
	}
	defer f.Close()

	newG := rdf2go.NewGraph(l.prefix, false)

	err = newG.Parse(f, "text/turtle")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse file %s", pathToFile)
	}

	var ret []*rdf2go.Triple
	for s := range newG.IterTriples() {
		ret = append(ret, s)
	}

	return ret, nil
}
