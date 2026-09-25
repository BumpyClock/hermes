package parser

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/BumpyClock/hermes/internal/definitions"
	"github.com/BumpyClock/hermes/internal/resource"
)

const definitionGoldenPath = "testdata/canonical_definition_output.json"

type definitionGolden struct {
	InputsSHA256 string            `json:"inputs_sha256"`
	Results      map[string]string `json:"results"`
}

// TestCanonicalDefinitionOutputMatchesGolden pins every canonical fixture's
// public Result digest so content pipeline refactors stay byte-identical.
func TestCanonicalDefinitionOutputMatchesGolden(t *testing.T) {
	root := os.Getenv("HERMES_DEFINITIONS_PILOT")
	if root == "" {
		t.Skip("set HERMES_DEFINITIONS_PILOT to compare canonical definition fixtures with the golden output")
	}
	snapshot, err := definitions.LoadDirectory(filepath.Join(root, "definitions"))
	if err != nil {
		t.Fatal(err)
	}
	inputs := sha256.New()
	definitionFiles, err := filepath.Glob(filepath.Join(root, "definitions", "*.y*ml"))
	if err != nil {
		t.Fatal(err)
	}
	caseFiles, err := filepath.Glob(filepath.Join(root, "fixtures", "*", "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(definitionFiles)
	sort.Strings(caseFiles)
	hashFile := func(path string) []byte {
		//nolint:gosec // The operator selects this external synthetic-fixture repository.
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		rel, _ := filepath.Rel(root, path)
		sum := sha256.Sum256(data)
		inputs.Write([]byte(filepath.ToSlash(rel) + "\x00" + hex.EncodeToString(sum[:]) + "\n"))
		return data
	}
	for _, path := range definitionFiles {
		hashFile(path)
	}

	got := definitionGolden{Results: map[string]string{}}
	for _, caseFile := range caseFiles {
		var fixture struct {
			Cases []struct{ File, URL string }
		}
		if err = json.Unmarshal(hashFile(caseFile), &fixture); err != nil {
			t.Fatal(err)
		}
		for _, c := range fixture.Cases {
			page := hashFile(filepath.Join(filepath.Dir(caseFile), c.File))
			parsedURL, parseErr := url.Parse(c.URL)
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			rule := snapshot.Match(parsedURL.Hostname())
			if rule == nil {
				t.Fatalf("%s: no definition matches %s", caseFile, c.URL)
			}
			for _, contentType := range []string{"html", "markdown", "text"} {
				result := parseCanonicalFixture(t, snapshot, string(page), c.URL, parsedURL, contentType)
				if result.ExtractorUsed != "definition:"+rule.Domain {
					t.Fatalf("%s %s: extractor %q, want definition:%s", c.File, contentType, result.ExtractorUsed, rule.Domain)
				}
				encoded, encodeErr := json.Marshal(result)
				if encodeErr != nil {
					t.Fatal(encodeErr)
				}
				sum := sha256.Sum256(encoded)
				rel, _ := filepath.Rel(root, filepath.Join(filepath.Dir(caseFile), c.File))
				got.Results[filepath.ToSlash(rel)+" "+contentType] = hex.EncodeToString(sum[:])
			}
		}
	}
	got.InputsSHA256 = hex.EncodeToString(inputs.Sum(nil))

	if os.Getenv("HERMES_UPDATE_DEFINITION_GOLDEN") != "" {
		encoded, encodeErr := json.MarshalIndent(got, "", "  ")
		if encodeErr != nil {
			t.Fatal(encodeErr)
		}
		//nolint:gosec // The golden is a checked-in test fixture.
		if err = os.WriteFile(definitionGoldenPath, append(encoded, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	raw, err := os.ReadFile(definitionGoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	var want definitionGolden
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if got.InputsSHA256 != want.InputsSHA256 {
		t.Fatalf("definition inputs changed (%s, golden %s); regenerate the golden at an unchanged engine commit with HERMES_UPDATE_DEFINITION_GOLDEN=1",
			got.InputsSHA256, want.InputsSHA256)
	}
	for key, digest := range want.Results {
		if got.Results[key] != digest {
			t.Errorf("%s: result digest %s, golden %s", key, got.Results[key], digest)
		}
	}
	if len(got.Results) != len(want.Results) {
		t.Errorf("%d results, golden has %d", len(got.Results), len(want.Results))
	}
}

// parseCanonicalFixture runs ParseHTMLWithContext minus DNS validation, which
// would need the network.
func parseCanonicalFixture(t *testing.T, snapshot *definitions.Snapshot, page, targetURL string, parsedURL *url.URL, contentType string) *Result {
	t.Helper()
	doc, err := resource.CreateDocument(context.Background(), targetURL, page, parsedURL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := New().extractAllFieldsWithContext(context.Background(), doc, targetURL, parsedURL, ParserOptions{
		Definitions: snapshot, Fallback: true, ContentType: contentType,
	})
	if err != nil {
		t.Fatalf("%s %s: %v", targetURL, contentType, err)
	}
	return result
}
