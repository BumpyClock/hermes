package hermes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoadManagedDefinitionsPinnedReleaseAndCacheFallback(t *testing.T) {
	manifest, archive := managedBundle(t)
	cache := managedTestCache(t)
	transport := managedReleaseTransport(t, "synthetic-1", manifest, archive)

	loaded, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache, Transport: transport,
	})
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Snapshot == nil || loaded.Version != "synthetic-1" || loaded.Digest == "" ||
		loaded.Source != "managed-release" || loaded.CacheStatus != "activated" || !loaded.Updated || loaded.Warning != nil {
		t.Fatalf("unexpected fresh managed outcome: %+v", loaded)
	}

	fallback, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedTransport(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("upstream unavailable")
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if fallback.Snapshot == nil || fallback.Source != "managed-cache" || fallback.Warning == nil || fallback.Updated {
		t.Fatalf("expected warned cached fallback, got %+v", fallback)
	}
}

func TestLoadManagedDefinitionsAutomaticSelectsChronologicalCompatibleRelease(t *testing.T) {
	baseManifest, archive := managedBundle(t)
	oldManifest := automaticManifest(t, baseManifest, "definition-old")
	newManifest := automaticManifest(t, baseManifest, "definition-new")
	incompatible := bytes.Replace(newManifest, []byte(`"metadata.text"`), []byte(`"unknown.aaaaa"`), 1)
	releases := []automaticRelease{
		{tag: "future-prerelease", publishedAt: "2026-01-04T00:00:00Z", prerelease: true, manifest: newManifest, archive: archive},
		{tag: "incompatible", publishedAt: "2026-01-03T00:00:00Z", manifest: incompatible, archive: archive},
		{tag: "release-old", publishedAt: "2026-01-01T00:00:00Z", manifest: oldManifest, archive: archive},
		{tag: "release-new", publishedAt: "2026-01-02T00:00:00Z", manifest: newManifest, archive: archive},
	}
	cache := managedTestCache(t)
	transport := managedAutomaticTransport(t, releases)
	loaded, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Automatic: true, CacheDirectory: cache, Transport: transport,
	})
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != "definition-new" || loaded.Source != "managed-release" || loaded.UpdateStatus != "update-new" || !loaded.Updated {
		t.Fatalf("automatic activation = %+v", loaded)
	}

	current, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Automatic: true, CacheDirectory: cache, Transport: transport,
	})
	if err != nil {
		t.Fatal(err)
	}
	if current.Version != "definition-new" || current.Source != "managed-cache" || current.UpdateStatus != "cache-current" || current.Updated || current.Warning != nil {
		t.Fatalf("automatic current cache = %+v", current)
	}

	fallback, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Automatic: true, CacheDirectory: cache,
		Transport: managedTransport(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("release inventory unavailable")
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if fallback.Version != "definition-new" || fallback.UpdateStatus != "cache-after-failure" || fallback.Warning == nil {
		t.Fatalf("automatic cache fallback = %+v", fallback)
	}

	brokenManifest := automaticManifest(t, baseManifest, "definition-broken")
	brokenArchive := append([]byte(nil), archive...)
	brokenArchive[len(brokenArchive)-1] = 1
	invalidCandidate, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Automatic: true, CacheDirectory: cache,
		Transport: managedAutomaticTransport(t, []automaticRelease{
			{tag: "release-broken", publishedAt: "2026-01-05T00:00:00Z", manifest: brokenManifest, archive: brokenArchive},
			{tag: "release-new", publishedAt: "2026-01-02T00:00:00Z", manifest: newManifest, archive: archive},
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if invalidCandidate.Version != "definition-new" || invalidCandidate.UpdateStatus != "cache-after-failure" || invalidCandidate.Warning == nil {
		t.Fatalf("invalid automatic candidate fallback = %+v", invalidCandidate)
	}
}

func TestLoadManagedDefinitionsAutomaticPaginationAndPinnedRollback(t *testing.T) {
	baseManifest, archive := managedBundle(t)
	oldManifest := automaticManifest(t, baseManifest, "definition-old")
	newManifest := automaticManifest(t, baseManifest, "definition-new")
	cache := managedTestCache(t)
	oldTransport := managedReleaseTransport(t, "definition-old", oldManifest, archive)
	if _, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "definition-old", CacheDirectory: cache, Transport: oldTransport,
	}); err != nil {
		t.Fatal(err)
	}

	releases := make([]automaticRelease, 0, automaticReleasePageSize+1)
	for i := 0; i < automaticReleasePageSize; i++ {
		releases = append(releases, automaticRelease{
			tag: fmt.Sprintf("draft-%02d", i), publishedAt: "2026-01-03T00:00:00Z", draft: true,
		})
	}
	releases = append(releases, automaticRelease{
		tag: "release-new", publishedAt: "2026-01-02T00:00:00Z", manifest: newManifest, archive: archive,
	})
	if _, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Automatic: true, CacheDirectory: cache, Transport: managedAutomaticTransport(t, releases),
	}); err != nil {
		t.Fatal(err)
	}

	rolledBack, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "definition-old", CacheDirectory: cache,
		Transport: managedTransport(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("offline rollback")
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if rolledBack.Version != "definition-old" || rolledBack.Warning == nil || rolledBack.UpdateStatus != "cache-after-failure" {
		t.Fatalf("pinned rollback outcome = %+v", rolledBack)
	}
}

func TestLoadManagedDefinitionsAutomaticCacheUsesReleaseTagTieBreak(t *testing.T) {
	baseManifest, archive := managedBundle(t)
	manifestForATag := automaticManifest(t, baseManifest, "definition-z")
	manifestForZTag := automaticManifest(t, baseManifest, "definition-a")
	const publishedAt = "2026-01-02T00:00:00Z"
	cache := managedTestCache(t)

	if _, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Automatic: true, CacheDirectory: cache,
		Transport: managedAutomaticTransport(t, []automaticRelease{
			{tag: "a-release", publishedAt: publishedAt, manifest: manifestForATag, archive: archive},
		}),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Automatic: true, CacheDirectory: cache,
		Transport: managedAutomaticTransport(t, []automaticRelease{
			{tag: "z-release", publishedAt: publishedAt, manifest: manifestForZTag, archive: archive},
		}),
	}); err != nil {
		t.Fatal(err)
	}

	ordered, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Automatic: true, CacheDirectory: cache,
		Transport: managedAutomaticTransport(t, []automaticRelease{
			{tag: "a-release", publishedAt: publishedAt, manifest: manifestForATag, archive: archive},
			{tag: "z-release", publishedAt: publishedAt, manifest: manifestForZTag, archive: archive},
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if ordered.Version != "definition-a" || ordered.UpdateStatus != "cache-current" {
		t.Fatalf("live equal-time tag ordering = %+v", ordered)
	}

	fallback, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Automatic: true, CacheDirectory: cache,
		Transport: managedTransport(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("inventory outage")
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if fallback.Version != "definition-a" || fallback.UpdateStatus != "cache-after-failure" || fallback.Warning == nil {
		t.Fatalf("cached equal-time tag ordering = %+v", fallback)
	}
}

func TestLoadManagedDefinitionsCoordinatesAutomaticAndPinnedStartup(t *testing.T) {
	baseManifest, archive := managedBundle(t)
	oldManifest := automaticManifest(t, baseManifest, "definition-old")
	newManifest := automaticManifest(t, baseManifest, "definition-new")
	automatic := managedAutomaticTransport(t, []automaticRelease{
		{tag: "release-new", publishedAt: "2026-01-02T00:00:00Z", manifest: newManifest, archive: archive},
	})
	pinned := managedReleaseTransport(t, "definition-old", oldManifest, archive)
	transport := managedTransport(func(request *http.Request) (*http.Response, error) {
		if strings.Contains(request.URL.String(), "definition-old") {
			return pinned.RoundTrip(request)
		}
		return automatic.RoundTrip(request)
	})
	cache := managedTestCache(t)
	type result struct {
		loaded *ManagedDefinitions
		err    error
	}
	results := make(chan result, 2)
	for _, options := range []ManagedDefinitionsOptions{
		{Automatic: true, CacheDirectory: cache, Transport: transport},
		{Version: "definition-old", CacheDirectory: cache, Transport: transport},
	} {
		go func(options ManagedDefinitionsOptions) {
			loaded, err := LoadManagedDefinitions(context.Background(), options)
			results <- result{loaded: loaded, err: err}
		}(options)
	}
	versions := map[string]bool{}
	for range 2 {
		outcome := <-results
		if outcome.err != nil {
			t.Fatal(outcome.err)
		}
		versions[outcome.loaded.Version] = true
	}
	if !versions["definition-new"] || !versions["definition-old"] {
		t.Fatalf("coordinated startup versions = %v", versions)
	}
}

func TestLoadManagedDefinitionsAutomaticRejectsConflictingMode(t *testing.T) {
	_, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "definition-old", Automatic: true, CacheDirectory: managedTestCache(t),
	})
	if err == nil || !strings.Contains(err.Error(), "cannot specify") {
		t.Fatalf("conflicting managed mode = %v", err)
	}
}

func TestLoadManagedDefinitionsPinnedRejectsDescriptorAndManifestMismatches(t *testing.T) {
	manifest, archive := managedBundle(t)
	cache := managedTestCache(t)
	descriptorMismatch := managedTransport(func(request *http.Request) (*http.Response, error) {
		if strings.Contains(request.URL.Path, "/releases/tags/synthetic-1") {
			body := `{"tag_name":"other-release","draft":false,"prerelease":false,"assets":[]}`
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
		}
		return nil, fmt.Errorf("unexpected request %s", request.URL)
	})
	_, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache, Transport: descriptorMismatch,
	})
	if err == nil || !strings.Contains(err.Error(), "requested stable release") {
		t.Fatalf("descriptor tag mismatch = %v", err)
	}

	wrongManifest := automaticManifest(t, manifest, "other-version")
	_, err = LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedReleaseTransport(t, "synthetic-1", wrongManifest, archive),
	})
	if err == nil || !strings.Contains(err.Error(), "does not match pin") {
		t.Fatalf("manifest version mismatch = %v", err)
	}
}

func TestLoadManagedDefinitionsUsesCacheForAcquisitionDeadlines(t *testing.T) {
	manifest, archive := managedBundle(t)
	cache := managedTestCache(t)
	if _, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedReleaseTransport(t, "synthetic-1", manifest, archive),
	}); err != nil {
		t.Fatal(err)
	}

	for name, transport := range map[string]http.RoundTripper{
		"request": managedTransport(func(*http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		}),
		"body": managedDeadlineBodyTransport(t, "synthetic-1", manifest, archive),
	} {
		t.Run(name, func(t *testing.T) {
			loaded, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
				Version: "synthetic-1", CacheDirectory: cache, Transport: transport,
			})
			if err != nil {
				t.Fatal(err)
			}
			if loaded.Snapshot == nil || loaded.Warning == nil || !errors.Is(loaded.Warning, context.DeadlineExceeded) {
				t.Fatalf("expected warned cached timeout fallback, got %+v", loaded)
			}
		})
	}

	_, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: managedTestCache(t),
		Transport: managedTransport(func(*http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		}),
	})
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cold-cache timeout = %v, want fatal deadline error", err)
	}
}

func TestLoadManagedDefinitionsCallerCancellationDoesNotUseCache(t *testing.T) {
	manifest, archive := managedBundle(t)
	cache := managedTestCache(t)
	if _, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedReleaseTransport(t, "synthetic-1", manifest, archive),
	}); err != nil {
		t.Fatal(err)
	}
	for name, cancelled := range map[string]context.Context{
		"cancelled": func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		"deadline": func() context.Context {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			defer cancel()
			return ctx
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := LoadManagedDefinitions(cancelled, ManagedDefinitionsOptions{
				Version: "synthetic-1", CacheDirectory: cache,
				Transport: managedTransport(func(*http.Request) (*http.Response, error) {
					return nil, errors.New("must not request after caller cancellation")
				}),
			})
			if !errors.Is(err, cancelled.Err()) {
				t.Fatalf("caller cancellation = %v, want %v", err, cancelled.Err())
			}
		})
	}

	t.Run("during-discovery", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err := LoadManagedDefinitions(ctx, ManagedDefinitionsOptions{
			Version: "synthetic-1", CacheDirectory: cache,
			Transport: managedTransport(func(*http.Request) (*http.Response, error) {
				cancel()
				return nil, context.Canceled
			}),
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("mid-request cancellation = %v, want context.Canceled", err)
		}
	})
}

func TestLoadManagedDefinitionsRedactsSignedRedirects(t *testing.T) {
	manifest, archive := managedBundle(t)
	cache := managedTestCache(t)
	if _, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedReleaseTransport(t, "synthetic-1", manifest, archive),
	}); err != nil {
		t.Fatal(err)
	}

	const secret = "signed-token-must-not-leak"
	manifestURL := "https://github.com/BumpyClock/hermes-definitions/releases/download/synthetic-1/manifest.json"
	base := managedReleaseTransport(t, "synthetic-1", manifest, archive)
	var rejectedHostCalls atomic.Int32
	redirecting := managedTransport(func(request *http.Request) (*http.Response, error) {
		if request.URL.Hostname() == "untrusted.example" {
			rejectedHostCalls.Add(1)
			return nil, errors.New("redirect target was requested")
		}
		if request.URL.String() == manifestURL {
			return &http.Response{
				StatusCode: http.StatusFound,
				Header:     http.Header{"Location": []string{"https://user:" + secret + "@untrusted.example/private?token=" + secret + "#fragment"}},
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    request,
			}, nil
		}
		return base.RoundTrip(request)
	})
	loaded, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache, Transport: redirecting,
	})
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Warning == nil || strings.Contains(loaded.Warning.Error(), secret) || !strings.Contains(loaded.Warning.Error(), "untrusted.example/private") {
		t.Fatalf("unsafe redirect warning = %v", loaded.Warning)
	}
	var leakedURL *url.Error
	if errors.As(loaded.Warning, &leakedURL) {
		t.Fatalf("redirect warning exposes URL-bearing cause: %v", leakedURL)
	}
	if rejectedHostCalls.Load() != 0 {
		t.Fatal("unsafe redirect destination reached the transport")
	}

	_, err = LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: managedTestCache(t), Transport: redirecting,
	})
	if err == nil || strings.Contains(err.Error(), secret) || !strings.Contains(err.Error(), "untrusted.example/private") {
		t.Fatalf("unsafe redirect error = %v", err)
	}
	if errors.As(err, &leakedURL) {
		t.Fatalf("redirect error exposes URL-bearing cause: %v", leakedURL)
	}
}

func TestLoadManagedDefinitionsRejectsChangedKnownIdentity(t *testing.T) {
	manifest, archive := managedBundle(t)
	cache := managedTestCache(t)
	if _, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedReleaseTransport(t, "synthetic-1", manifest, archive),
	}); err != nil {
		t.Fatal(err)
	}

	changedDigest := strings.Repeat("a", 64)
	changedManifest := bytes.ReplaceAll(manifest, []byte(`7e3ac7726f3a07bc60d6795798cd781d3b7dae90e4a7f8515ecd841131a4b3b2`), []byte(changedDigest))
	_, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedReleaseTransport(t, "synthetic-1", changedManifest, archive),
	})
	if err == nil || !strings.Contains(err.Error(), "changed known archive identity") {
		t.Fatalf("expected changed identity rejection, got %v", err)
	}
}

func TestLoadManagedDefinitionsCancellationAndInvalidArchive(t *testing.T) {
	manifest, archive := managedBundle(t)
	cache := managedTestCache(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := LoadManagedDefinitions(ctx, ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedReleaseTransport(t, "synthetic-1", manifest, archive),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation = %v, want context.Canceled", err)
	}

	corrupt := append([]byte(nil), archive...)
	corrupt[len(corrupt)-1] = 1
	_, err = LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedReleaseTransport(t, "synthetic-1", manifest, corrupt),
	})
	if err == nil || !strings.Contains(err.Error(), "release archive") {
		t.Fatalf("expected archive rejection, got %v", err)
	}
}

func TestLoadManagedDefinitionsRejectsIncompatibleAndCorruptCache(t *testing.T) {
	manifest, archive := managedBundle(t)
	unsupported := bytes.Replace(manifest, []byte(`"metadata.text"`), []byte(`"unknown.aaaaa"`), 1)
	_, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: managedTestCache(t),
		Transport: managedReleaseTransport(t, "synthetic-1", unsupported, archive),
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported required operation") {
		t.Fatalf("expected compatibility rejection, got %v", err)
	}

	cache := managedTestCache(t)
	if _, err = LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedReleaseTransport(t, "synthetic-1", manifest, archive),
	}); err != nil {
		t.Fatal(err)
	}
	digest := digestFromManifest(t, manifest)
	if err = os.WriteFile(filepath.Join(cache, "snapshots", "synthetic-1", digest, "bundle.tar"), []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
		Version: "synthetic-1", CacheDirectory: cache,
		Transport: managedTransport(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("rate limited")
		}),
	})
	if err == nil || !strings.Contains(err.Error(), "no usable cached snapshot") {
		t.Fatalf("expected corrupted cache fatal error, got %v", err)
	}
}

func TestLoadManagedDefinitionsConcurrentPinnedStartup(t *testing.T) {
	manifest, archive := managedBundle(t)
	cache := managedTestCache(t)
	var calls atomic.Int32
	transport := managedReleaseTransport(t, "synthetic-1", manifest, archive)
	counted := managedTransport(func(request *http.Request) (*http.Response, error) {
		calls.Add(1)
		return transport.RoundTrip(request)
	})
	const workers = 8
	results := make(chan *ManagedDefinitions, workers)
	failures := make(chan error, workers)
	var wait sync.WaitGroup
	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			loaded, err := LoadManagedDefinitions(context.Background(), ManagedDefinitionsOptions{
				Version: "synthetic-1", CacheDirectory: cache, Transport: counted,
			})
			if err != nil {
				failures <- err
				return
			}
			results <- loaded
		}()
	}
	wait.Wait()
	close(failures)
	close(results)
	for err := range failures {
		t.Error(err)
	}
	var digest string
	for loaded := range results {
		if loaded.Snapshot == nil || loaded.Digest == "" {
			t.Errorf("invalid concurrent outcome: %+v", loaded)
			continue
		}
		if digest == "" {
			digest = loaded.Digest
		} else if loaded.Digest != digest {
			t.Errorf("inconsistent digest %q, want %q", loaded.Digest, digest)
		}
	}
	if calls.Load() == 0 {
		t.Fatal("managed startup did not acquire the pinned release")
	}
	entries, err := os.ReadDir(filepath.Join(cache, "snapshots", "synthetic-1"))
	if err != nil || len(entries) != 1 || !entries[0].IsDir() {
		t.Fatalf("expected one atomically activated snapshot, entries=%v err=%v", entries, err)
	}
}

type managedTransport func(*http.Request) (*http.Response, error)

func (f managedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type automaticRelease struct {
	tag, publishedAt  string
	draft, prerelease bool
	manifest, archive []byte
}

const automaticReleasePageSize = 50

func managedAutomaticTransport(t *testing.T, releases []automaticRelease) http.RoundTripper {
	t.Helper()
	type asset struct {
		Name string `json:"name"`
		Size int    `json:"size"`
		URL  string `json:"browser_download_url"`
	}
	type descriptor struct {
		TagName     string  `json:"tag_name"`
		Draft       bool    `json:"draft"`
		Prerelease  bool    `json:"prerelease"`
		PublishedAt string  `json:"published_at"`
		Assets      []asset `json:"assets"`
	}
	descriptors := make([]descriptor, len(releases))
	bodies := map[string][]byte{}
	for i, release := range releases {
		descriptors[i] = descriptor{TagName: release.tag, Draft: release.draft, Prerelease: release.prerelease, PublishedAt: release.publishedAt}
		if release.manifest != nil {
			manifestURL := "https://github.com/BumpyClock/hermes-definitions/releases/download/" + release.tag + "/manifest.json"
			digest := digestFromManifest(t, release.manifest)
			archiveName := "definitions-" + manifestVersion(t, release.manifest) + "-" + digest + ".tar"
			archiveURL := "https://github.com/BumpyClock/hermes-definitions/releases/download/" + release.tag + "/" + archiveName
			descriptors[i].Assets = []asset{
				{Name: "manifest.json", Size: len(release.manifest), URL: manifestURL},
				{Name: archiveName, Size: len(release.archive), URL: archiveURL},
			}
			bodies[manifestURL], bodies[archiveURL] = release.manifest, release.archive
		}
	}
	return managedTransport(func(request *http.Request) (*http.Response, error) {
		var body []byte
		switch request.URL.String() {
		case "https://api.github.com/repos/BumpyClock/hermes-definitions/releases?per_page=50&page=1":
			var err error
			body, err = json.Marshal(descriptors[:min(automaticReleasePageSize, len(descriptors))])
			if err != nil {
				return nil, err
			}
		case "https://api.github.com/repos/BumpyClock/hermes-definitions/releases?per_page=50&page=2":
			var err error
			if len(descriptors) > automaticReleasePageSize {
				body, err = json.Marshal(descriptors[automaticReleasePageSize:])
			} else {
				body, err = json.Marshal([]descriptor{})
			}
			if err != nil {
				return nil, err
			}
		default:
			var ok bool
			body, ok = bodies[request.URL.String()]
			if !ok {
				return nil, fmt.Errorf("unexpected automatic managed request %s", request.URL)
			}
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(body)), Request: request}, nil
	})
}

func automaticManifest(t *testing.T, manifest []byte, version string) []byte {
	t.Helper()
	return bytes.ReplaceAll(manifest, []byte("synthetic-1"), []byte(version))
}

func manifestVersion(t *testing.T, manifest []byte) string {
	t.Helper()
	const prefix = `"version":"`
	start := bytes.Index(manifest, []byte(prefix))
	if start < 0 {
		t.Fatal("manifest has no version")
	}
	start += len(prefix)
	end := bytes.IndexByte(manifest[start:], '"')
	if end < 0 {
		t.Fatal("manifest version is unterminated")
	}
	return string(manifest[start : start+end])
}

type deadlineBody struct{}

func (deadlineBody) Read([]byte) (int, error) { return 0, context.DeadlineExceeded }

func managedDeadlineBodyTransport(t *testing.T, version string, manifest, archive []byte) http.RoundTripper {
	t.Helper()
	base := managedReleaseTransport(t, version, manifest, archive)
	manifestURL := "https://github.com/BumpyClock/hermes-definitions/releases/download/" + version + "/manifest.json"
	return managedTransport(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() == manifestURL {
			return &http.Response{
				StatusCode: http.StatusOK, Header: make(http.Header),
				Body: io.NopCloser(deadlineBody{}), Request: request,
			}, nil
		}
		return base.RoundTrip(request)
	})
}

func managedReleaseTransport(t *testing.T, version string, manifest, archive []byte) http.RoundTripper {
	t.Helper()
	manifestURL := "https://github.com/BumpyClock/hermes-definitions/releases/download/" + version + "/manifest.json"
	archiveName := "definitions-" + version + "-" + digestFromManifest(t, manifest) + ".tar"
	archiveURL := "https://github.com/BumpyClock/hermes-definitions/releases/download/" + version + "/" + archiveName
	descriptor := fmt.Sprintf(`{"tag_name":%q,"draft":false,"prerelease":false,"assets":[{"name":"manifest.json","size":%d,"browser_download_url":%q},{"name":%q,"size":%d,"browser_download_url":%q}]}`,
		version, len(manifest), manifestURL, archiveName, len(archive), archiveURL)
	return managedTransport(func(request *http.Request) (*http.Response, error) {
		var body []byte
		switch request.URL.String() {
		case "https://api.github.com/repos/BumpyClock/hermes-definitions/releases/tags/" + version:
			body = []byte(descriptor)
		case manifestURL:
			body = manifest
		case archiveURL:
			body = archive
		default:
			return nil, fmt.Errorf("unexpected managed request %s", request.URL)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    request,
		}, nil
	})
}

func digestFromManifest(t *testing.T, manifest []byte) string {
	t.Helper()
	const prefix = `"sha256":"`
	start := bytes.Index(manifest, []byte(prefix))
	if start < 0 {
		t.Fatal("manifest has no archive digest")
	}
	start += len(prefix)
	if len(manifest) < start+64 {
		t.Fatal("manifest archive digest is truncated")
	}
	return string(manifest[start : start+64])
}

func managedBundle(t *testing.T) ([]byte, []byte) {
	t.Helper()
	manifest, err := os.ReadFile("internal/definitionbundle/testdata/pinned/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile("internal/definitionbundle/testdata/pinned/bundle.tar")
	if err != nil {
		t.Fatal(err)
	}
	return manifest, archive
}

func managedTestCache(t *testing.T) string {
	t.Helper()
	cache, err := os.MkdirTemp(".", ".managed-definitions-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(cache) })
	return cache
}
