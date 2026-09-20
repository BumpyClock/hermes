package definitionbundle_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	bundle "github.com/BumpyClock/hermes/internal/definitionbundle"
)

func TestCaseFoldedPayloadCollisions(t *testing.T) {
	for _, paths := range [][2]string{
		{"definitions/A.yaml", "definitions/a.yaml"},
		{"fixtures/example/Article.html", "fixtures/example/article.html"},
		{"fixtures/EXAMPLE/article.html", "fixtures/example/article.html"},
	} {
		t.Run(paths[0], func(t *testing.T) {
			m, files := pinned(t)
			for i, name := range paths {
				data := []byte("distinct payload " + strings.Repeat("x", i+1))
				files[name] = data
			}
			m.Files = nil
			for name, data := range files {
				m.Files = append(m.Files, bundle.File{Path: name, Size: int64(len(data)), SHA256: bundle.Digest(data)})
			}
			slices.SortFunc(m.Files, func(a, b bundle.File) int { return strings.Compare(a.Path, b.Path) })
			archive := encodedArchive(t, m, files, nil)
			bindArchive(m, archive)
			wire, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = bundle.ParseManifest(wire); err == nil {
				t.Error("case-folded manifest aliases accepted")
			}
			if _, err = m.ReadArchive(archive, m.Archive.SHA256); err == nil {
				t.Error("case-folded archive aliases accepted")
			}
			dir := filepath.Join(t.TempDir(), "must-not-create")
			if err = bundle.WriteDefinitions(files, dir); err == nil {
				t.Error("case-folded payload aliases reached extraction")
			}
			if _, err = os.Stat(dir); !os.IsNotExist(err) {
				t.Error("collision rejection must precede directory creation")
			}
		})
	}
}

func TestExistingDestinationSentinelUnchanged(t *testing.T) {
	directory := t.TempDir()
	sentinel := filepath.Join(directory, "A.yaml")
	if err := os.WriteFile(sentinel, []byte("external sentinel"), 0o600); err != nil {
		t.Fatal(err)
	}

	if data, err := bundle.ReadFile(filepath.Join(directory, "a.yaml"), 1024); err == nil {
		t.Logf("case-insensitive filesystem observed: %s", data)
	} else {
		t.Log("case-sensitive filesystem; portable collision checks still apply")
	}
	if err := bundle.WriteDefinitions(map[string][]byte{"definitions/a.yaml": []byte("replacement")}, directory); err == nil {
		t.Fatal("existing extraction directory accepted")
	}
	data, err := bundle.ReadFile(sentinel, 1024)
	if err != nil || string(data) != "external sentinel" {
		t.Fatalf("external sentinel changed: %q %v", data, err)
	}
}

func TestArticleHostBindingMatchesRuntimeNormalization(t *testing.T) {
	for _, host := range []string{"www.example.com", "www.example.com.", "www.example.com.."} {
		t.Run(host, func(t *testing.T) {
			m, files := pinned(t)
			suite, err := bundle.ParseSuite(files)
			if err != nil {
				t.Fatal(err)
			}
			directory := filepath.Join(t.TempDir(), "definitions")
			if err = bundle.WriteDefinitions(files, directory); err != nil {
				t.Fatal(err)
			}
			suite.Cases[0].URL = "https://" + host + "/report"
			suite.Cases[0].Expect = map[string]string{"url": suite.Cases[0].URL}
			suite.Cases[0].ExactlyOnce, suite.Cases[0].Exclude = []string{}, []string{}
			err = m.AuditRequirements(files, suite, directory)
			if strings.HasSuffix(host, "..") {
				if err == nil || !strings.Contains(err.Error(), "does not select") {
					t.Fatalf("double-dot host must fail real definition binding: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWireKeysAreExactBeforeTypedLoad(t *testing.T) {
	m, files := pinned(t)
	wire, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	for _, replacement := range [][2]string{
		{`"protocol":1`, `"protocol":2,"Protocol":1`},
		{`"protocol":1`, `"Protocol":1`},
		{`"operations":`, `"operations":["unknown.operation"],"Operations":`},
		{`"algorithms":`, `"Algorithms":`},
		{`"revision":`, `"Revision":`},
		{`"sha256":`, `"SHA256":`},
		{`"path":`, `"Path":`},
	} {
		t.Run(replacement[1], func(t *testing.T) {
			data := bytes.Replace(wire, []byte(replacement[0]), []byte(replacement[1]), 1)
			if _, err := bundle.ParseManifest(data); err == nil {
				t.Fatalf("wire alias reached typed manifest loader: %s", replacement[1])
			}
		})
	}
	for _, replacement := range [][2]string{
		{`"protocol":1`, `"Protocol":1`},
		{`"definition":`, `"Definition":`},
		{`"synthetic":`, `"Synthetic":`},
		{`"exactly_once":`, `"Exactly_Once":`},
	} {
		candidate := map[string][]byte{}
		for name, data := range files {
			candidate[name] = data
		}
		candidate["conformance.json"] = bytes.Replace(files["conformance.json"], []byte(replacement[0]), []byte(replacement[1]), 1)
		if _, err := bundle.ParseSuite(candidate); err == nil {
			t.Fatalf("wire alias reached suite loader: %s", replacement[1])
		}
	}
	var free map[string]string
	if err := bundle.DecodeJSON([]byte(`{"Mixed":"first","mixed":"second"}`), &free); err != nil || len(free) != 2 {
		t.Fatalf("truly free-form map keys must remain case-sensitive and distinct: %v %v", free, err)
	}
}
