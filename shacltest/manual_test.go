package shacltest

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/utrack/shaclngo/bitgraph"
)

func TestManual(t *testing.T) {
	ownPrefix := "http://datashapes.org/sh/tests/core/node/and-001.ttl"
	dir := "../data-shapes/data-shapes-test-suite/tests/core/node/and-001.ttl"
	// ownPrefix := "http://datashapes.org/sh/tests/core/"
	// dir := "../data-shapes/data-shapes-test-suite/tests/core/manifest.ttl"
	g, err := loadGraph(ownPrefix, dir)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("pre-graph")
	graph, err := bitgraph.FromRDFGraph(g)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("post-graph")

	var buf bytes.Buffer
	err = g.Serialize(&buf, "text/turtle")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("getTestManifests from Test")

	_ = getTestManifests(graph)

	os.WriteFile("testmanual_result.ttl", buf.Bytes(), 0644)

}
