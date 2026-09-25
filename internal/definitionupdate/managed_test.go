package definitionupdate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/BumpyClock/hermes/internal/definitionbundle"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func okResponse(request *http.Request, body []byte) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK, Header: make(http.Header),
		Body: io.NopCloser(bytes.NewReader(body)), Request: request,
	}
}

func inventoryClient(t *testing.T, inventory string) *http.Client {
	t.Helper()
	firstPage := fmt.Sprintf("%s/releases?per_page=%d&page=1", repositoryAPI, releasePageSize)
	return newClient(roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != firstPage {
			return nil, fmt.Errorf("unexpected request %s", request.URL)
		}
		return okResponse(request, []byte(inventory)), nil
	}))
}

func TestFetchReleasesOrdersByParsedPublicationTime(t *testing.T) {
	// String order disagrees with instant order for "middle"/"newest" and for
	// the tie pair, which shares one instant and falls back to descending tag.
	inventory := `[
		{"tag_name":"old","published_at":"2026-01-01T00:00:00Z"},
		{"tag_name":"tie-a","published_at":"2026-01-02T01:00:00+01:00"},
		{"tag_name":"draft","draft":true,"published_at":"not a time"},
		{"tag_name":"newest","published_at":"2026-01-03T00:00:00-05:00"},
		{"tag_name":"tie-b","published_at":"2026-01-02T00:00:00Z"},
		{"tag_name":"prerelease","prerelease":true,"published_at":"2027-01-01T00:00:00Z"},
		{"tag_name":"middle","published_at":"2026-01-03T03:00:00Z"}
	]`
	releases, err := fetchReleases(context.Background(), inventoryClient(t, inventory))
	if err != nil {
		t.Fatal(err)
	}
	var tags []string
	for _, release := range releases {
		tags = append(tags, release.TagName)
	}
	if got, want := strings.Join(tags, ","), "newest,middle,tie-b,tie-a,old"; got != want {
		t.Fatalf("release order = %s, want %s", got, want)
	}
	if releases[0].PublishedAt != "2026-01-03T00:00:00-05:00" {
		t.Fatalf("sorted release lost its original publication time: %+v", releases[0])
	}
}

func TestFetchReleasesRejectsInvalidStablePublicationTime(t *testing.T) {
	_, err := fetchReleases(context.Background(), inventoryClient(t, `[{"tag_name":"bad","published_at":"2026-01-01"}]`))
	var acquisition *acquisitionError
	if !errors.As(err, &acquisition) || !strings.Contains(err.Error(), `stable release "bad" has invalid publication time`) {
		t.Fatalf("invalid publication time error = %v", err)
	}
}

func testCacheRoot(t *testing.T) string {
	t.Helper()
	root, err := os.MkdirTemp(".", ".definitionupdate-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	absolute, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	return absolute
}

func TestCacheNamesRejectCaseFoldedAliases(t *testing.T) {
	root := testCacheRoot(t)
	// Each name exists in one casing only, so the layout is valid on
	// case-insensitive filesystems.
	for _, directory := range []string{"snapshots/Alpha", "snapshots/Shared", "identities"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"Beta.json", "SHARED.json", "Gamma.txt", "Delta.JSON"} {
		if err := os.WriteFile(filepath.Join(root, "identities", name), []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	names, err := listCacheNames(root)
	if err != nil {
		t.Fatal(err)
	}
	for version, want := range map[string]string{
		"alpha":     `case-folded managed cache version collision: "alpha" and "Alpha"`,
		"Alpha":     "",
		"beta":      `case-folded managed cache version collision: "beta" and "Beta"`,
		"Beta":      "",
		"shared":    `case-folded managed cache version collision: "shared" and "Shared"`,
		"gamma":     "",
		"delta":     "",
		"unrelated": "",
	} {
		err := names.rejectAlias(version)
		if (want == "" && err != nil) || (want != "" && (err == nil || err.Error() != want)) {
			t.Errorf("rejectAlias(%q) = %v, want %q", version, err, want)
		}
	}
}

func TestLoadRejectsCaseFoldedCacheVersionAlias(t *testing.T) {
	manifestBytes, err := os.ReadFile("../definitionbundle/testdata/pinned/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile("../definitionbundle/testdata/pinned/bundle.tar")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := definitionbundle.ParseManifest(manifestBytes)
	if err != nil {
		t.Fatal(err)
	}
	version := manifest.Version
	alias := strings.ToUpper(version)
	collision := fmt.Sprintf("case-folded managed cache version collision: %q and %q", version, alias)
	download := repositoryWeb + "/releases/download/" + version + "/"
	descriptor := fmt.Sprintf(`{"tag_name":%q,"assets":[{"name":"manifest.json","size":%d,"browser_download_url":%q},{"name":%q,"size":%d,"browser_download_url":%q}]}`,
		version, len(manifestBytes), download+"manifest.json", manifest.Archive.Name, len(archive), download+manifest.Archive.Name)
	support := Support{Schema: manifest.DefinitionSchema, Operations: manifest.Engine.Operations, Algorithms: manifest.Engine.Algorithms}

	for _, test := range []struct {
		name string
		// seed runs before Load; onArchive runs when the archive is requested.
		seed, onArchive func(root string) error
		offline         bool
		wantArchive     bool
	}{
		{
			name: "before download",
			seed: func(root string) error { return os.MkdirAll(filepath.Join(root, "snapshots", alias), 0o700) },
		},
		{
			name:        "during download",
			onArchive:   func(root string) error { return os.Mkdir(filepath.Join(root, "snapshots", alias), 0o700) },
			wantArchive: true,
		},
		{
			name: "cache fallback",
			seed: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "identities"), 0o700); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(root, "identities", alias+".json"), []byte("{}"), 0o600)
			},
			offline: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := testCacheRoot(t)
			if test.seed != nil {
				if err := test.seed(root); err != nil {
					t.Fatal(err)
				}
			}
			archiveRequested := false
			transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if test.offline {
					return nil, errors.New("offline")
				}
				switch request.URL.String() {
				case repositoryAPI + "/releases/tags/" + version:
					return okResponse(request, []byte(descriptor)), nil
				case download + "manifest.json":
					return okResponse(request, manifestBytes), nil
				case download + manifest.Archive.Name:
					archiveRequested = true
					if test.onArchive != nil {
						if err := test.onArchive(root); err != nil {
							return nil, err
						}
					}
					return okResponse(request, archive), nil
				}
				return nil, fmt.Errorf("unexpected request %s", request.URL)
			})
			_, _, err := Load(context.Background(), Config{Version: version, CacheDirectory: root, Transport: transport, Support: support})
			if err == nil || !strings.Contains(err.Error(), collision) {
				t.Fatalf("Load error = %v, want %q", err, collision)
			}
			if archiveRequested != test.wantArchive {
				t.Fatalf("archive requested = %v, want %v", archiveRequested, test.wantArchive)
			}
			entries, err := os.ReadDir(filepath.Join(root, "identities"))
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if entry.Name() == version+".json" {
					t.Fatalf("aliased version recorded an identity")
				}
			}
			if test.onArchive != nil {
				activated, err := os.ReadDir(filepath.Join(root, "snapshots", alias))
				if err != nil || len(activated) != 0 {
					t.Fatalf("aliased version activated a snapshot: entries=%v err=%v", activated, err)
				}
			}
		})
	}
}

func TestPrepareRootCreatesOnlyCacheDirectories(t *testing.T) {
	root, err := prepareRoot(testCacheRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if got := strings.Join(names, ","); got != "identities,locks,snapshots" {
		t.Fatalf("cache root entries = %s", got)
	}
}

// Validation reads a stored snapshot without writing, so it cannot depend on
// the snapshot directory's depth below the cache root.
func TestLoadSnapshotValidatesInMemoryAtAnyDepth(t *testing.T) {
	manifestBytes, err := os.ReadFile("../definitionbundle/testdata/pinned/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile("../definitionbundle/testdata/pinned/bundle.tar")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := definitionbundle.ParseManifest(manifestBytes)
	if err != nil {
		t.Fatal(err)
	}
	support := Support{Schema: manifest.DefinitionSchema, Operations: manifest.Engine.Operations, Algorithms: manifest.Engine.Algorithms}
	for _, parts := range [][]string{{manifest.Archive.SHA256}, {"a", "b", "c", "d", manifest.Archive.SHA256}} {
		t.Run(fmt.Sprint(len(parts)), func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(append([]string{root}, parts...)...)
			if err := os.MkdirAll(directory, 0o700); err != nil {
				t.Fatal(err)
			}
			for name, data := range map[string][]byte{"manifest.json": manifestBytes, "bundle.tar": archive} {
				if err := os.WriteFile(filepath.Join(directory, name), data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			before := treeEntries(t, root)
			snapshot, metadata, err := loadSnapshot(directory, support)
			if err != nil || snapshot == nil || metadata.Digest != manifest.Archive.SHA256 {
				t.Fatalf("loadSnapshot = %v, %+v, %v", snapshot, metadata, err)
			}
			if after := treeEntries(t, root); !slices.Equal(before, after) {
				t.Fatalf("validation changed the filesystem:\n before %v\n after  %v", before, after)
			}
		})
	}
}

func treeEntries(t *testing.T, root string) []string {
	t.Helper()
	var entries []string
	err := filepath.WalkDir(root, func(path string, _ os.DirEntry, err error) error {
		entries = append(entries, path)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return entries
}
