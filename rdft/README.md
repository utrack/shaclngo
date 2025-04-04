# RDFT: RDF Triple to Go Unmarshaller

RDFT is a Go library for unmarshalling RDF triples (from Turtle and other RDF formats) into Go structs. It provides a flexible and type-safe way to work with RDF data in Go applications.

## Features

- Unmarshal RDF triples into Go structs using struct tags
- Support for all RDF value types (resources, literals, blank nodes)
- Handle localized strings with language tags
- Support for collections of values (slices)
- Type conversion between RDF literals and Go types
- Namespace handling for compact URIs
- Custom unmarshalling via the `RDFUnmarshaler` interface

## Installation

```bash
go get github.com/utrack/shaclngo/rdft
```

## Basic Usage

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/deiu/rdf2go"
    "github.com/utrack/shaclngo/rdft"
)

// Define a struct that maps to your RDF data
type Person struct {
    URI       string   `rdf:"@id"`
    Type      string   `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
    Name      string   `rdf:"http://xmlns.com/foaf/0.1/name"`
    Age       int      `rdf:"http://xmlns.com/foaf/0.1/age"`
    Knows     []string `rdf:"http://xmlns.com/foaf/0.1/knows"`
    Biography rdft.LocalizedText `rdf:"http://xmlns.com/foaf/0.1/bio" rdf-type:"localized"`
}

func main() {
    // Load RDF data
    g := rdf2go.NewGraph("http://example.org/", false)
    f, _ := os.Open("data.ttl")
    g.Parse(f, "text/turtle")
    
    // Create an unmarshaller
    u := rdft.NewUnmarshaller(g)
    
    // Unmarshal a resource
    var person Person
    err := u.Unmarshal("http://example.org/person/1", &person)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Person: %+v\n", person)
    
    // Get a localized biography in a specific language
    englishBio := person.Biography.GetWithFallback("en", "")
    fmt.Printf("English bio: %s\n", englishBio)
}
```

## Struct Tags

RDFT uses struct tags to map Go struct fields to RDF predicates:

- `rdf:"<predicate>"`: Specifies the RDF predicate URI for the field
- `rdf-type:"localized"`: Indicates that the field contains localized text (with language tags)

Special predicates:
- `@id`: Maps to the resource URI itself

## Type Mapping

RDFT automatically converts between RDF literals and Go types:

| RDF Type | Go Type |
|----------|---------|
| xsd:string | string |
| xsd:integer | int, int64 |
| xsd:decimal | float64 |
| xsd:boolean | bool |
| rdf:langString | rdft.LocalizedString, rdft.LocalizedText |
| Resource URI | string, rdft.Resource |
| Blank Node | string, rdft.BlankNode |

## Working with Localized Text

RDFT provides two types for working with language-tagged strings:

1. `LocalizedString`: For a single string with a language tag
   ```go
   type Article struct {
       Title rdft.LocalizedString `rdf:"http://example.org/title" rdf-type:"localized"`
   }
   ```

2. `LocalizedText`: For multiple translations of the same text
   ```go
   type Article struct {
       Description rdft.LocalizedText `rdf:"http://example.org/description" rdf-type:"localized"`
   }
   
   // Get text in a specific language
   englishDesc := article.Description.Get("en")
   
   // Get text with fallback language
   desc := article.Description.GetWithFallback("de", "en")
   ```

## Custom Unmarshalling

You can implement the `RDFUnmarshaler` interface to customize how a type is unmarshalled:

```go
type CustomType struct {
    // fields
}

func (c *CustomType) UnmarshalRDF(values []rdft.Value) error {
    // Custom unmarshalling logic
    return nil
}
```

## Namespace Handling

RDFT provides utilities for working with RDF namespaces:

```go
// Create a namespace map with common namespaces
ns := rdft.CommonNamespaces()

// Add a custom namespace
ns.AddNamespace("ex", "http://example.org/")

// Expand a qualified name
uri, _ := ns.ExpandQName("foaf:Person")
// uri = "http://xmlns.com/foaf/0.1/Person"

// Compact a URI
qname := ns.CompactURI("http://xmlns.com/foaf/0.1/name")
// qname = "foaf:name"
```

## License

MIT License
