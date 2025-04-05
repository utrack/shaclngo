package shacltest

import (
	"bytes"
	"fmt"
	"os"
	"testing"
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

	fmt.Println("graph loaded")

	var buf bytes.Buffer
	err = g.Serialize(&buf, "text/turtle")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("GetTestManifests from Test")

	tests, err := GetTestManifests(g)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Found %d tests\n", len(tests))

	os.WriteFile("testmanual_result.ttl", buf.Bytes(), 0644)

}
