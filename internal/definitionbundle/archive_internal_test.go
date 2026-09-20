package definitionbundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExclusiveDestinationCreation(t *testing.T) {
	directory := t.TempDir()
	sentinel := filepath.Join(directory, "sentinel")
	if err := os.WriteFile(sentinel, []byte("external sentinel"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(directory, "link.yaml")
	if err := os.Symlink(sentinel, link); err != nil {
		t.Fatal(err)
	}
	for _, destination := range []string{sentinel, link} {
		if err := writeNewFile(destination, []byte("replacement")); !os.IsExist(err) {
			t.Fatalf("existing destination not rejected exclusively: %v", err)
		}
	}
	data, err := ReadFile(sentinel, 1024)
	if err != nil || string(data) != "external sentinel" {
		t.Fatalf("exclusive creation modified sentinel: %q %v", data, err)
	}
}
