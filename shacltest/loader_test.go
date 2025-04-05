package shacltest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseManifest(t *testing.T) {
	so := require.New(t)
	ownPrefix := "http://datashapes.org/sh/tests/core/node/and-001.ttl#"
	dir := "../data-shapes/data-shapes-test-suite/tests/core/node/and-001.ttl"
	// ownPrefix := "http://datashapes.org/sh/tests/core/"
	// dir := "../data-shapes/data-shapes-test-suite/tests/core/manifest.ttl"
	g, err := loadGraph(ownPrefix, dir)
	so.NoError(err)

	manifest, err := GetTestManifests(g)
	so.NoError(err)

	so.Equal(1, len(manifest.Entries))

	test := manifest.Entries[0]
	so.Equal("http://datashapes.org/sh/tests/core/node/and-001.ttl#and-001", test.ID)
	so.Equal("Test of sh:and at node shape 001", test.Description)
	so.NotNil(test.Status)
	so.NotNil(test.Action.Shapes)
	so.NotNil(test.Action.Data)
	so.NotNil(test.ExpectedResult)

	so.Equal(false, test.ExpectedResult.Conforms)
	so.Len(test.ExpectedResult.Results, 2)
}
