package hermes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCanonicalDefinitionDirectoryLoads(t *testing.T) {
	root := os.Getenv("HERMES_DEFINITIONS_PILOT")
	if root == "" {
		t.Skip("set HERMES_DEFINITIONS_PILOT to load the explicit canonical local definition directory")
	}
	snapshot, err := LoadDefinitions(filepath.Join(root, "definitions"))
	if err != nil {
		t.Fatal(err)
	}
	for _, host := range []string{"bbc.com", "www.nytimes.com", "www.theverge.com", "www.abendblatt.de"} {
		if snapshot.snapshot.Match(host) == nil {
			t.Fatalf("canonical directory has no definition for %q", host)
		}
	}
}
