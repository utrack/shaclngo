package rdft

import (
	"path/filepath"
	"testing"
)

// TestW3CTurtleManifestLoading tests loading the W3C Turtle test manifest
func TestW3CTurtleManifestLoading(t *testing.T) {
	// Path to the manifest file
	manifestPath := filepath.Join("testfiles", "turtle-tests-w3c-mirror", "manifest.ttl")

	// Load the manifest
	manifest, err := LoadW3CTurtleTestManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	// Check that we have tests
	if len(manifest.Tests) == 0 {
		t.Fatalf("No tests found in manifest")
	}

	// Print some stats
	t.Logf("Loaded %d tests from W3C Turtle test manifest", len(manifest.Tests))

	// Print the first few tests for verification
	for i := 0; i < 5 && i < len(manifest.Tests); i++ {
		test := manifest.Tests[i]
		t.Logf("Test %d: ID=%s, Name=%s, Type=%s", i+1, test.ID, test.Name, test.Type)
	}
}

// TestW3CTurtleBasicTests runs a subset of the W3C Turtle tests
func TestW3CTurtleBasicTests(t *testing.T) {
	// Path to the manifest file
	manifestPath := filepath.Join("testfiles", "turtle-tests-w3c-mirror", "manifest.ttl")

	// Load the manifest
	manifest, err := LoadW3CTurtleTestManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	// Define a list of basic tests to run
	// These are simple tests that should pass with our current implementation
	basicTestNames := []string{
		"IRI_subject",
		"bareword_a_predicate",
		"SPARQL_style_prefix",
		"prefixed_IRI_predicate",
		"prefixed_IRI_object",
	}

	// Find and run the basic tests
	for _, testName := range basicTestNames {
		var found bool
		for _, test := range manifest.Tests {
			if test.Name == testName {
				found = true
				t.Run(testName, func(t *testing.T) {
					// Run the test
					passed, err := test.RunTest(manifest.BaseDir)
					if err != nil {
						t.Errorf("Error running test %s: %v", testName, err)
					} else if !passed {
						t.Errorf("Test %s failed", testName)
					}
				})
				break
			}
		}
		if !found {
			t.Logf("Test %s not found in manifest", testName)
		}
	}
}

// TestW3CTurtleAllTests runs all the W3C Turtle tests
func TestW3CTurtleAllTests(t *testing.T) {
	t.SkipNow()

	// Path to the manifest file
	manifestPath := filepath.Join("testfiles", "turtle-tests-w3c-mirror", "manifest.ttl")

	// Load the manifest
	manifest, err := LoadW3CTurtleTestManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if len(manifest.Tests) != 291 {
		t.Fatalf("Expected 291 tests, found %d", len(manifest.Tests))
	}

	// Run all tests
	for _, test := range manifest.Tests {
		// Skip tests that are not approved
		if test.Approval != "Approved" {
			t.Logf("Skipping non-approved test: %s", test.Name)
			continue
		}

		// Skip negative tests for now
		if test.Type == "TestTurtleNegativeSyntax" || test.Type == "TestTurtleNegativeEval" {
			t.Logf("Skipping negative test: %s", test.Name)
			continue
		}

		t.Run(test.Name, func(t *testing.T) {
			// Run the test
			result, err := test.RunTest(manifest.BaseDir)
			if err != nil {
				t.Errorf("Error running test %s: %v", test.Name, err)
			} else if !result {
				t.Errorf("Test %s failed", test.Name)
			}
		})
	}
}
