package hermes

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BumpyClock/hermes/internal/definitionbundle"
)

func TestManagedDefinitionsCandidate(t *testing.T) {
	directory := os.Getenv("HERMES_DEFINITION_CANDIDATE")
	if directory == "" {
		t.Skip("set HERMES_DEFINITION_CANDIDATE and HERMES_DEFINITION_MANIFEST_SHA256 to verify a prepared local bundle")
	}
	manifestBytes, err := os.ReadFile(filepath.Join(directory, "manifest.json")) //nolint:gosec // Explicit local candidate path; content is pinned below.
	if err != nil {
		t.Fatal(err)
	}
	if definitionbundle.Digest(manifestBytes) != os.Getenv("HERMES_DEFINITION_MANIFEST_SHA256") {
		t.Fatal("candidate manifest does not match the explicit SHA-256 pin")
	}
	manifest, err := definitionbundle.ParseManifest(manifestBytes)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile(filepath.Join(directory, manifest.Archive.Name)) //nolint:gosec // Strict manifest validates the asset basename and pins its bytes.
	if err != nil {
		t.Fatal(err)
	}
	files, err := manifest.ReadArchive(archive, manifest.Archive.SHA256)
	if err != nil {
		t.Fatal(err)
	}
	suite, err := definitionbundle.ParseSuite(files)
	if err != nil {
		t.Fatal(err)
	}
	localDirectory := filepath.Join(t.TempDir(), "definitions")
	if err = definitionbundle.WriteDefinitions(files, localDirectory); err != nil {
		t.Fatal(err)
	}
	local, err := LoadDefinitions(localDirectory)
	if err != nil {
		t.Fatal(err)
	}
	definitionCount := 0
	for path := range files {
		if strings.HasPrefix(path, "definitions/") {
			definitionCount++
		}
	}
	if definitionCount != 125 {
		t.Fatalf("full migration candidate needs 125 definitions, got %d", definitionCount)
	}
	for _, automatic := range []bool{false, true} {
		t.Run(map[bool]string{false: "pinned", true: "automatic"}[automatic], func(t *testing.T) {
			options := ManagedDefinitionsOptions{
				Version: manifest.Version, CacheDirectory: managedTestCache(t),
				Transport: managedReleaseTransport(t, manifest.Version, manifestBytes, archive),
			}
			if automatic {
				options.Version, options.Automatic = "", true
				options.Transport = managedAutomaticTransport(t, []automaticRelease{{
					tag: manifest.Version, publishedAt: "2026-09-20T00:00:00Z",
					manifest: manifestBytes, archive: archive,
				}})
			}
			loaded, loadErr := LoadManagedDefinitions(context.Background(), options)
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			if loaded.Warning != nil || loaded.Version != manifest.Version || loaded.Digest != manifest.Archive.SHA256 {
				t.Fatalf("unexpected candidate identity: %+v", loaded)
			}
			checkCandidateMatches(t, suite, local, loaded.Snapshot)
			options.Transport = managedTransport(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("deterministic offline acquisition failure")
			})
			cached, loadErr := LoadManagedDefinitions(context.Background(), options)
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			if cached.Warning == nil || cached.UpdateStatus != "cache-after-failure" || cached.Digest != loaded.Digest {
				t.Fatalf("expected warning with identical validated cache: %+v", cached)
			}
			checkCandidateMatches(t, suite, local, cached.Snapshot)
			options.CacheDirectory = managedTestCache(t)
			if missing, missingErr := LoadManagedDefinitions(context.Background(), options); missingErr == nil || missing != nil {
				t.Fatal("offline acquisition without a cache must fail")
			}
		})
	}
	t.Logf("validated %d definitions and %d case host bindings through local, pinned, automatic, and cached sources", definitionCount, len(suite.Cases))
}

func checkCandidateMatches(t *testing.T, suite *definitionbundle.Suite, local, managed *Definitions) {
	t.Helper()
	if managed == nil || managed.snapshot == nil {
		t.Fatal("managed success returned no immutable snapshot")
	}
	for _, test := range suite.Cases {
		target, err := url.Parse(test.URL)
		if err != nil {
			t.Fatal(err)
		}
		want, got := local.snapshot.Match(target.Hostname()), managed.snapshot.Match(target.Hostname())
		if want == nil || got == nil || want.Domain != got.Domain {
			t.Fatalf("managed host binding differs for case %s", test.ID)
		}
	}
}
