package rdft

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/deiu/rdf2go"
)

// Namespace represents an RDF namespace
type Namespace struct {
	Prefix string
	URI    string
}

// NewNamespace creates a new namespace with the given prefix and URI
func NewNamespace(prefix, uri string) *Namespace {
	return &Namespace{
		Prefix: prefix,
		URI:    uri,
	}
}

// ExpandQName expands a qualified name using the given namespace
func (ns *Namespace) ExpandQName(qname string) string {
	if strings.HasPrefix(qname, ns.Prefix+":") {
		localName := strings.TrimPrefix(qname, ns.Prefix+":")
		return ns.URI + localName
	}
	return qname
}

// CompactURI compacts a URI using the given namespace
func (ns *Namespace) CompactURI(uri string) string {
	if strings.HasPrefix(uri, ns.URI) {
		localName := strings.TrimPrefix(uri, ns.URI)
		return ns.Prefix + ":" + localName
	}
	return uri
}

// NamespaceMap is a collection of namespaces
type NamespaceMap struct {
	namespaces map[string]*Namespace // prefix -> namespace
	uriMap     map[string]*Namespace // uri -> namespace
}

// NewNamespaceMap creates a new empty namespace map
func NewNamespaceMap() *NamespaceMap {
	return &NamespaceMap{
		namespaces: make(map[string]*Namespace),
		uriMap:     make(map[string]*Namespace),
	}
}

// Add adds a namespace to the map
func (nm *NamespaceMap) Add(ns *Namespace) {
	nm.namespaces[ns.Prefix] = ns
	nm.uriMap[ns.URI] = ns
}

// AddNamespace adds a namespace with the given prefix and URI
func (nm *NamespaceMap) AddNamespace(prefix, uri string) {
	ns := NewNamespace(prefix, uri)
	nm.Add(ns)
}

// GetByPrefix returns a namespace by its prefix
func (nm *NamespaceMap) GetByPrefix(prefix string) (*Namespace, bool) {
	ns, ok := nm.namespaces[prefix]
	return ns, ok
}

// GetByURI returns a namespace by its URI
func (nm *NamespaceMap) GetByURI(uri string) (*Namespace, bool) {
	ns, ok := nm.uriMap[uri]
	return ns, ok
}

// ExpandQName expands a qualified name using the appropriate namespace
func (nm *NamespaceMap) ExpandQName(qname string) (string, error) {
	parts := strings.SplitN(qname, ":", 2)
	if len(parts) != 2 {
		return qname, nil // Not a QName
	}
	
	prefix := parts[0]
	localName := parts[1]
	
	ns, ok := nm.GetByPrefix(prefix)
	if !ok {
		return "", fmt.Errorf("unknown prefix: %s", prefix)
	}
	
	return ns.URI + localName, nil
}

// CompactURI compacts a URI using the appropriate namespace
func (nm *NamespaceMap) CompactURI(uri string) string {
	for _, ns := range nm.namespaces {
		if strings.HasPrefix(uri, ns.URI) {
			localName := strings.TrimPrefix(uri, ns.URI)
			return ns.Prefix + ":" + localName
		}
	}
	return uri // No matching namespace
}

// CommonNamespaces returns a namespace map with common RDF namespaces
func CommonNamespaces() *NamespaceMap {
	nm := NewNamespaceMap()
	
	nm.AddNamespace("rdf", "http://www.w3.org/1999/02/22-rdf-syntax-ns#")
	nm.AddNamespace("rdfs", "http://www.w3.org/2000/01/rdf-schema#")
	nm.AddNamespace("xsd", "http://www.w3.org/2001/XMLSchema#")
	nm.AddNamespace("owl", "http://www.w3.org/2002/07/owl#")
	nm.AddNamespace("sh", "http://www.w3.org/ns/shacl#")
	nm.AddNamespace("dc", "http://purl.org/dc/elements/1.1/")
	nm.AddNamespace("dcterms", "http://purl.org/dc/terms/")
	nm.AddNamespace("foaf", "http://xmlns.com/foaf/0.1/")
	nm.AddNamespace("skos", "http://www.w3.org/2004/02/skos/core#")
	
	return nm
}

// ExtractNamespaces extracts namespaces from an RDF graph
func ExtractNamespaces(graph *rdf2go.Graph) *NamespaceMap {
	nm := NewNamespaceMap()
	
	// Add common namespaces
	for prefix, ns := range CommonNamespaces().namespaces {
		nm.AddNamespace(prefix, ns.URI)
	}
	
	// Extract namespaces from the graph
	for triple := range graph.IterTriples() {
		// Extract from subject
		if res, ok := triple.Subject.(*rdf2go.Resource); ok {
			extractNamespaceFromURI(nm, res.URI)
		}
		
		// Extract from predicate
		if res, ok := triple.Predicate.(*rdf2go.Resource); ok {
			extractNamespaceFromURI(nm, res.URI)
		}
		
		// Extract from object
		if res, ok := triple.Object.(*rdf2go.Resource); ok {
			extractNamespaceFromURI(nm, res.URI)
		}
	}
	
	return nm
}

// extractNamespaceFromURI attempts to extract a namespace from a URI
func extractNamespaceFromURI(nm *NamespaceMap, uri string) {
	// Skip if already in a known namespace
	for _, ns := range nm.namespaces {
		if strings.HasPrefix(uri, ns.URI) {
			return
		}
	}
	
	// Try to extract a namespace
	parts := strings.Split(uri, "#")
	if len(parts) == 2 && parts[1] != "" {
		// URI with fragment
		nsURI := parts[0] + "#"
		prefix := guessPrefix(nsURI)
		nm.AddNamespace(prefix, nsURI)
		return
	}
	
	// Try with last path segment
	lastSlash := strings.LastIndex(uri, "/")
	if lastSlash != -1 && lastSlash < len(uri)-1 {
		nsURI := uri[:lastSlash+1]
		prefix := guessPrefix(nsURI)
		nm.AddNamespace(prefix, nsURI)
	}
}

// guessPrefix tries to guess a reasonable prefix for a namespace URI
func guessPrefix(uri string) string {
	// Remove common prefixes
	uri = strings.TrimPrefix(uri, "http://")
	uri = strings.TrimPrefix(uri, "https://")
	
	// Remove common suffixes
	uri = strings.TrimSuffix(uri, "#")
	uri = strings.TrimSuffix(uri, "/")
	
	// Split by dots and slashes
	parts := strings.FieldsFunc(uri, func(r rune) bool {
		return r == '.' || r == '/'
	})
	
	if len(parts) == 0 {
		return "ns"
	}
	
	// Use the last meaningful part
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "www" && parts[i] != "com" && parts[i] != "org" && parts[i] != "net" {
			return strings.ToLower(parts[i])
		}
	}
	
	return "ns"
}

// GetRDFType returns the RDF type of a Go value
func GetRDFType(v interface{}) string {
	if v == nil {
		return ""
	}
	
	val := reflect.ValueOf(v)
	typ := val.Type()
	
	switch typ.Kind() {
	case reflect.String:
		return "http://www.w3.org/2001/XMLSchema#string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "http://www.w3.org/2001/XMLSchema#integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "http://www.w3.org/2001/XMLSchema#nonNegativeInteger"
	case reflect.Float32, reflect.Float64:
		return "http://www.w3.org/2001/XMLSchema#decimal"
	case reflect.Bool:
		return "http://www.w3.org/2001/XMLSchema#boolean"
	case reflect.Struct:
		if typ == reflect.TypeOf(LocalizedString{}) {
			return "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"
		}
	}
	
	return ""
}
