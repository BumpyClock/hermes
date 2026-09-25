package parser

import (
	"bytes"
	"encoding/json"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/BumpyClock/hermes/internal/definitions"
)

// TestCanonicalDefinitionsLoadIdenticallyFromMemory checks that the in-memory
// loader used for managed releases yields the same public Result as the
// directory loader for every canonical fixture.
func TestCanonicalDefinitionsLoadIdenticallyFromMemory(t *testing.T) {
	root := os.Getenv("HERMES_DEFINITIONS_PILOT")
	if root == "" {
		t.Skip("set HERMES_DEFINITIONS_PILOT to compare in-memory and directory loading of canonical definitions")
	}
	directory := filepath.Join(root, "definitions")
	fromDir, err := definitions.LoadDirectory(directory)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		//nolint:gosec // The operator selects this external synthetic-fixture repository.
		data, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		files[entry.Name()] = data
	}
	fromMemory, err := definitions.LoadFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(fromDir.Sites(), fromMemory.Sites()) {
		t.Fatalf("sites differ: directory %v, memory %v", fromDir.Sites(), fromMemory.Sites())
	}
	dirOperations, dirAlgorithms := fromDir.UsedCapabilities()
	memoryOperations, memoryAlgorithms := fromMemory.UsedCapabilities()
	if !slices.Equal(dirOperations, memoryOperations) || !slices.Equal(dirAlgorithms, memoryAlgorithms) {
		t.Fatalf("capabilities differ: directory %v %v, memory %v %v", dirOperations, dirAlgorithms, memoryOperations, memoryAlgorithms)
	}

	caseFiles, err := filepath.Glob(filepath.Join(root, "fixtures", "*", "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	compared := 0
	for _, caseFile := range caseFiles {
		//nolint:gosec // The operator selects this external synthetic-fixture repository.
		raw, readErr := os.ReadFile(caseFile)
		if readErr != nil {
			t.Fatal(readErr)
		}
		var fixture struct {
			Cases []struct{ File, URL string }
		}
		if err = json.Unmarshal(raw, &fixture); err != nil {
			t.Fatal(err)
		}
		for _, c := range fixture.Cases {
			//nolint:gosec // The operator selects this external synthetic-fixture repository.
			page, pageErr := os.ReadFile(filepath.Join(filepath.Dir(caseFile), c.File))
			if pageErr != nil {
				t.Fatal(pageErr)
			}
			parsedURL, parseErr := url.Parse(c.URL)
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			for _, contentType := range []string{"html", "markdown", "text"} {
				want, wantErr := json.Marshal(parseCanonicalFixture(t, fromDir, string(page), c.URL, parsedURL, contentType))
				got, gotErr := json.Marshal(parseCanonicalFixture(t, fromMemory, string(page), c.URL, parsedURL, contentType))
				if wantErr != nil || gotErr != nil {
					t.Fatalf("encode results: %v, %v", wantErr, gotErr)
				}
				if !bytes.Equal(got, want) {
					t.Errorf("%s %s: in-memory definitions changed the result", c.File, contentType)
				}
				compared++
			}
		}
	}
	if compared == 0 {
		t.Fatal("no canonical fixture cases compared")
	}
}
