package rgraph

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/deiu/rdf2go"
	"github.com/utrack/caisson-go/errors"
)

type Graph interface {
	One(subject, predicate, object rdf2go.Term) *rdf2go.Triple
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
	if subject == nil && predicate == nil && object == nil {
		var ret []*rdf2go.Triple
		for t := range g.g.IterTriples() {
			ret = append(ret, t)
		}
		return ret
	}
	return g.g.All(subject, predicate, object)
}

func (g *rdf2goGraph) One(subject, predicate, object rdf2go.Term) *rdf2go.Triple {
	return g.g.One(subject, predicate, object)
}

func (g *rdf2goGraph) LoadURI(uri string) error {
	ts, err := g.l.LoadURI(uri)
	if err != nil {
		return err
	}
	
	// Add all triples to the graph
	for _, t := range ts {
		g.g.AddTriple(t.Subject, t.Predicate, t.Object)
	}
	
	// Create a resource for the file URI
	fileResource := rdf2go.NewResource(uri)
	
	// Create the rdfs:isDefinedBy predicate
	isDefinedByPredicate := rdf2go.NewResource("http://www.w3.org/2000/01/rdf-schema#isDefinedBy")
	
	// Track unique subjects to avoid duplicates
	processedSubjects := make(map[string]bool)
	
	// Find all named resources (subjects) in the loaded triples
	for _, t := range ts {
		// Only process resources (not blank nodes or literals)
		resource, isResource := t.Subject.(*rdf2go.Resource)
		if !isResource {
			continue
		}
		
		// Skip if we've already processed this subject
		subjectStr := resource.String()
		if _, alreadyProcessed := processedSubjects[subjectStr]; alreadyProcessed {
			continue
		}
		
		// Check if this resource already has any rdfs:isDefinedBy triple
		existingDefinitions := g.g.All(resource, isDefinedByPredicate, nil)
		if len(existingDefinitions) > 0 {
			// Already has an isDefinedBy link, so skip
			processedSubjects[subjectStr] = true
			continue
		}
		
		// Add the rdfs:isDefinedBy link
		g.g.AddTriple(resource, isDefinedByPredicate, fileResource)
		processedSubjects[subjectStr] = true
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
