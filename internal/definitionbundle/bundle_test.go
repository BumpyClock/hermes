package definitionbundle_test

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	bundle "github.com/BumpyClock/hermes/internal/definitionbundle"
	"github.com/BumpyClock/hermes/internal/definitions"

	hermes "github.com/BumpyClock/hermes"
)

func pinned(t *testing.T) (*bundle.Manifest, map[string][]byte) {
	t.Helper()
	manifestData, err := os.ReadFile("testdata/pinned/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	pinsData, err := os.ReadFile("testdata/pinned/pins.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins map[string]string
	if err = bundle.DecodeJSON(pinsData, &pins); err != nil {
		t.Fatal(err)
	}
	if bundle.Digest(manifestData) != pins["manifest_sha256"] {
		t.Fatal("prepared manifest pin mismatch")
	}
	m, err := bundle.ParseManifest(manifestData)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile("testdata/pinned/bundle.tar")
	if err != nil {
		t.Fatal(err)
	}
	files, err := m.ReadArchive(archive, pins["archive_sha256"])
	if err != nil {
		t.Fatal(err)
	}
	return m, files
}

func load(t *testing.T, directory string) *definitions.Snapshot {
	t.Helper()
	snapshot, err := definitions.LoadDirectory(directory)
	if err != nil {
		t.Fatalf("real loader rejected definitions: %v", err)
	}
	return snapshot
}

func TestPreparedPinnedBundle(t *testing.T) {
	m, files := pinned(t)
	suite, err := bundle.ParseSuite(files)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "definitions")
	if err := bundle.WriteDefinitions(files, dir); err != nil {
		t.Fatal(err)
	}
	snapshot := load(t, dir)
	support := hermes.DefinitionCapabilities()
	if err := m.CheckSupport(support.Schema, support.Capabilities, []string{}); err != nil {
		t.Fatal(err)
	}
	if err := m.AuditRequirements(suite, snapshot); err != nil {
		t.Fatal(err)
	}
	if err := m.CheckCoverage(files, suite, nil); err != nil {
		t.Fatal(err)
	}
	if err := bundle.WriteDefinitions(files, dir); err == nil {
		t.Fatal("reused work directory accepted")
	}
}

func TestManifestRejectsMalformedContents(t *testing.T) {
	m, _ := pinned(t)
	base, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(*bundle.Manifest){
		"schema":      func(m *bundle.Manifest) { m.DefinitionSchema = 2 },
		"protocol":    func(m *bundle.Manifest) { m.Protocol = 2 },
		"scope":       func(m *bundle.Manifest) { m.Scope = "production-ish" },
		"contract":    func(m *bundle.Manifest) { m.Engine.Contract = "other" },
		"revision":    func(m *bundle.Manifest) { m.Source.Revision = "latest" },
		"identity":    func(m *bundle.Manifest) { m.Archive.Name = "different.tar" },
		"size":        func(m *bundle.Manifest) { m.Archive.Size = bundle.MaxArchiveSize + 1 },
		"format":      func(m *bundle.Manifest) { m.Archive.Format = "zip" },
		"cap-order":   func(m *bundle.Manifest) { slices.Reverse(m.Engine.Operations) },
		"cap-missing": func(m *bundle.Manifest) { m.Engine.Operations = nil },
		"cap-dupe":    func(m *bundle.Manifest) { m.Engine.Operations = []string{"a", "a"} },
		"traversal":   func(m *bundle.Manifest) { m.Files[0].Path = "../escape.yaml" },
		"absolute":    func(m *bundle.Manifest) { m.Files[0].Path = "/definitions/a.yaml" },
		"backslash":   func(m *bundle.Manifest) { m.Files[0].Path = `definitions\a.yaml` },
		"hidden":      func(m *bundle.Manifest) { m.Files[0].Path = "definitions/.a.yaml" },
		"file-order":  func(m *bundle.Manifest) { slices.Reverse(m.Files) },
		"payload":     func(m *bundle.Manifest) { m.Files[0].Size = bundle.MaxFileSize + 1 },
		"extension":   func(m *bundle.Manifest) { m.Files[0].Path = "run.sh" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			var candidate bundle.Manifest
			if err := json.Unmarshal(base, &candidate); err != nil {
				t.Fatal(err)
			}
			mutate(&candidate)
			data, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := bundle.ParseManifest(data); err == nil {
				t.Fatal("invalid manifest accepted")
			}
		})
	}
	for _, data := range [][]byte{
		bytes.Replace(base, []byte(`"protocol":1`), []byte(`"protocol":true`), 1),
		bytes.Replace(base, []byte(`"protocol":1`), []byte(`"protocol":"1"`), 1),
		bytes.Replace(base, []byte(`"protocol":1`), []byte(`"protocol":1,"protocol":1`), 1),
		bytes.Replace(base, []byte(`"protocol":1`), []byte(`"protocol":1,"extra":false`), 1),
		bytes.Replace(base, []byte(`"algorithms":[]`), []byte(`"algorithms":null`), 1),
		bytes.Replace(base, []byte(`"algorithms":[]`), []byte(`"algorithms":[true]`), 1),
		append(bytes.Clone(base), []byte("{}")...),
		[]byte("{\"invalid\":\"\xff\"}"),
		bytes.Repeat([]byte(" "), bundle.MaxManifestSize+1),
	} {
		if _, err := bundle.ParseManifest(data); err == nil {
			t.Fatal("invalid JSON envelope accepted")
		}
	}
}

func encodedArchive(t *testing.T, m *bundle.Manifest, files map[string][]byte, mutate func(*tar.Header)) []byte {
	t.Helper()
	var output bytes.Buffer
	writer := tar.NewWriter(&output)
	for _, file := range m.Files {
		header := &tar.Header{Name: file.Path, Mode: 0o644, Size: int64(len(files[file.Path])),
			Typeflag: tar.TypeReg, ModTime: time.Unix(0, 0), Format: tar.FormatUSTAR}
		if mutate != nil {
			mutate(header)
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if header.Typeflag == tar.TypeReg {
			if _, err := writer.Write(files[file.Path]); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func bindArchive(m *bundle.Manifest, data []byte) {
	m.Archive.Size = int64(len(data))
	m.Archive.SHA256 = bundle.Digest(data)
	m.Archive.Name = "definitions-" + m.Version + "-" + m.Archive.SHA256 + ".tar"
}

func TestArchiveRejectsUnsafeOrMismatchedBytes(t *testing.T) {
	for _, name := range []string{"traversal", "link", "mtime", "owner", "mode", "extra", "padding", "truncated", "file-digest", "pin"} {
		t.Run(name, func(t *testing.T) {
			m, files := pinned(t)
			data := encodedArchive(t, m, files, func(h *tar.Header) {
				switch name {
				case "traversal":
					h.Name = "../escape"
				case "link":
					h.Typeflag, h.Linkname, h.Size = tar.TypeSymlink, "/etc/passwd", 0
				case "mtime":
					h.ModTime = time.Unix(1, 0)
				case "owner":
					h.Uid = 1
				case "mode":
					h.Mode = 0o755
				}
			})
			switch name {
			case "extra":
				data = append(data, make([]byte, 512)...)
			case "padding":
				data[512+m.Files[0].Size] = 1
			case "truncated":
				data = data[:len(data)-512]
			case "file-digest":
				m.Files[0].SHA256 = strings.Repeat("0", 64)
			}
			bindArchive(m, data)
			pin := m.Archive.SHA256
			if name == "pin" {
				pin = strings.Repeat("0", 64)
			}
			if _, err := m.ReadArchive(data, pin); err == nil {
				t.Fatalf("unsafe archive accepted: %s", name)
			}
		})
	}
}

func TestCapabilityAndCaseFailures(t *testing.T) {
	m, files := pinned(t)
	suite, err := bundle.ParseSuite(files)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "defs")
	if err := bundle.WriteDefinitions(files, dir); err != nil {
		t.Fatal(err)
	}
	snapshot := load(t, dir)
	m.Engine.Operations = slices.DeleteFunc(m.Engine.Operations, func(s string) bool { return s == "metadata.text" })
	if err := m.AuditRequirements(suite, snapshot); err == nil || !strings.Contains(err.Error(), "metadata.text") {
		t.Fatalf("missing capability accepted: %v", err)
	}
	m.Engine.Operations = append(m.Engine.Operations, "unknown.operation")
	if err := m.CheckSupport(1, []string{}, []string{}); err == nil {
		t.Fatal("unknown operation accepted with matching schema")
	}
	m.Engine.Operations, m.Engine.Algorithms = []string{}, []string{"unknown.algorithm"}
	if err := m.CheckSupport(1, []string{}, []string{}); err == nil || !strings.Contains(err.Error(), "algorithm") {
		t.Fatalf("unknown algorithm accepted: %v", err)
	}
	m.Engine.Operations, m.Engine.Algorithms = hermes.DefinitionCapabilities().Capabilities, []string{}
	suite.Cases[0].URL = "https://wrong.example/report"
	if err := m.AuditRequirements(suite, snapshot); err == nil || !strings.Contains(err.Error(), "does not select") {
		t.Fatalf("wrong definition selection not rejected by real matcher: %v", err)
	}
	c := suite.Cases[0]
	failures := c.Compare(map[string]string{"title": "Wrong title", "content": "Generic decoy"}, "html")
	if len(failures) != 3 {
		t.Fatalf("expected every field/content mismatch, got %v", failures)
	}
	delete(files, c.Fixture)
	if _, err := bundle.ParseSuite(files); err == nil {
		t.Fatal("missing fixture accepted")
	}
}

func TestCompleteGateRequiresExactApprovedEvidence(t *testing.T) {
	m, files := pinned(t)
	suite, err := bundle.ParseSuite(files)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.CheckCoverage(files, suite, &bundle.ReleaseGate{}); err == nil {
		t.Fatal("pilot certified as production")
	}
	m.Scope = "complete"
	if err = m.CheckCoverage(files, suite, nil); err == nil {
		t.Fatal("complete bundle accepted without inventory/cohort gate")
	}
	gate := &bundle.ReleaseGate{Approved: true, InventorySHA256: strings.Repeat("a", 64),
		EvidenceSHA256: strings.Repeat("b", 64), Members: map[string]string{}, Sites: []bundle.CoverageSite{}}
	files = map[string][]byte{}
	suite = &bundle.Suite{Protocol: bundle.Protocol}
	real := false
	m.Files = nil
	for i := 0; i < 125; i++ {
		id := fmt.Sprintf("site%03d", i)
		definition := "definitions/" + id + ".yaml"
		gate.Members[id] = "cohort"
		gate.Sites = append(gate.Sites, bundle.CoverageSite{ID: id, Cohort: "cohort", Definition: definition, Cases: []string{id}})
		suite.Cases = append(suite.Cases, bundle.Case{ID: id, Definition: definition, Synthetic: &real})
		files[definition] = []byte("synthetic gate test")
		m.Files = append(m.Files, bundle.File{Path: definition, SHA256: bundle.Digest(files[definition]), Size: int64(len(files[definition]))})
	}
	recordBytes, err := json.Marshal(m.Files)
	if err != nil {
		t.Fatal(err)
	}
	gate.PayloadSHA256 = bundle.Digest(append(recordBytes, '\n'))
	coverage := bundle.Coverage{Protocol: bundle.Protocol, InventorySHA256: gate.InventorySHA256,
		EvidenceSHA256: gate.EvidenceSHA256, PayloadSHA256: gate.PayloadSHA256, Sites: gate.Sites}
	files["coverage.json"], err = json.Marshal(coverage)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.CheckCoverage(files, suite, gate); err != nil {
		t.Fatalf("exact 125-member approved gate rejected: %v", err)
	}
	gate.PayloadSHA256 = strings.Repeat("c", 64)
	if err := m.CheckCoverage(files, suite, gate); err == nil {
		t.Fatal("changed approved payload accepted")
	}
	gate.PayloadSHA256 = coverage.PayloadSHA256
	fake := true
	suite.Cases[0].Synthetic = &fake
	if err := m.CheckCoverage(files, suite, gate); err == nil {
		t.Fatal("synthetic fixture certified as fresh evidence")
	}
	suite.Cases[0].Synthetic = &real
	delete(gate.Members, "site000")
	if err := m.CheckCoverage(files, suite, gate); err == nil {
		t.Fatal("incomplete inventory accepted")
	}
}
