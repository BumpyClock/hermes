// Package definitionupdate loads immutable definition releases into an
// updater-owned cache, using exact pins or compatible stable discovery.
package definitionupdate

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
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/BumpyClock/hermes/internal/definitionbundle"
	"github.com/BumpyClock/hermes/internal/definitions"
)

const (
	repositoryAPI           = "https://api.github.com/repos/BumpyClock/hermes-definitions"
	repositoryWeb           = "https://github.com/BumpyClock/hermes-definitions"
	releasePageSize         = 50
	maxReleasePages         = 4
	maxReleaseInventorySize = 4 << 20
)

var (
	versionPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,95}$`)
	digestPattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Support describes the engine contract supplied by the public API boundary.
type Support struct {
	Schema     int
	Operations []string
	Algorithms []string
}

// Config controls one pinned or automatically selected release acquisition.
type Config struct {
	Version        string
	Automatic      bool
	CacheDirectory string
	Transport      http.RoundTripper
	Support        Support
}

// Metadata identifies the immutable release that was loaded.
type Metadata struct {
	Version      string
	Digest       string
	Source       string
	CacheStatus  string
	UpdateStatus string
	Updated      bool
	Warning      error
}

type release struct {
	TagName     string  `json:"tag_name"`
	Draft       bool    `json:"draft"`
	Prerelease  bool    `json:"prerelease"`
	PublishedAt string  `json:"published_at"`
	Assets      []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type identity struct {
	Version     string `json:"version"`
	Digest      string `json:"digest"`
	ReleaseTag  string `json:"release_tag"`
	PublishedAt string `json:"published_at"`
}

type acquisitionError struct{ err error }

func (e *acquisitionError) Error() string { return e.err.Error() }
func (e *acquisitionError) Unwrap() error { return e.err }

// Load selects a release according to config. It revalidates cache bytes and
// engine support before every success, including acquisition-failure fallback.
func Load(ctx context.Context, config Config) (*definitions.Snapshot, Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, Metadata{}, err
	}
	if config.Automatic && config.Version != "" {
		return nil, Metadata{}, fmt.Errorf("managed definition automatic mode cannot specify an exact version")
	}
	if !config.Automatic && !versionPattern.MatchString(config.Version) {
		return nil, Metadata{}, fmt.Errorf("managed definition version must be an exact release identifier")
	}
	if config.CacheDirectory == "" {
		return nil, Metadata{}, fmt.Errorf("managed definition cache directory is required")
	}
	if err := validateSupport(config.Support); err != nil {
		return nil, Metadata{}, err
	}
	root, err := prepareRoot(config.CacheDirectory)
	if err != nil {
		return nil, Metadata{}, err
	}

	unlock, err := acquireLock(ctx, filepath.Join(root, "locks", "managed-definitions.lock"))
	if err != nil {
		return nil, Metadata{}, err
	}
	defer func() { _ = unlock() }()

	if config.Automatic {
		return loadAutomatic(ctx, root, config)
	}
	return loadPinned(ctx, root, config)
}

func loadPinned(ctx context.Context, root string, config Config) (*definitions.Snapshot, Metadata, error) {
	snapshot, metadata, err := acquire(ctx, root, config)
	if err == nil {
		return snapshot, metadata, nil
	}
	if ctx.Err() != nil {
		return nil, Metadata{}, ctx.Err()
	}
	var acquisition *acquisitionError
	if !errors.As(err, &acquisition) {
		return nil, Metadata{}, fmt.Errorf("managed definitions %q: %w", config.Version, err)
	}
	cached, cacheMetadata, cacheErr := loadCachedVersion(root, config.Version, config.Support)
	if cacheErr != nil {
		return nil, Metadata{}, fmt.Errorf("managed definitions %q: acquisition failed (%w); no usable cached snapshot (%v)", config.Version, err, cacheErr)
	}
	cacheMetadata.Warning = fmt.Errorf("managed definitions update for %q failed; using validated cached snapshot: %w", config.Version, acquisition)
	cacheMetadata.UpdateStatus = "cache-after-failure"
	return cached, cacheMetadata, nil
}

func validateSupport(support Support) error {
	if support.Schema <= 0 || len(support.Operations) == 0 {
		return fmt.Errorf("managed definition engine support is incomplete")
	}
	if !slices.IsSorted(support.Operations) || !slices.IsSorted(support.Algorithms) {
		return fmt.Errorf("managed definition engine capabilities must be sorted")
	}
	return nil
}

func acquire(ctx context.Context, root string, config Config) (*definitions.Snapshot, Metadata, error) {
	client := newClient(config.Transport)
	descriptor, err := fetchRelease(ctx, client, config.Version)
	if err != nil {
		return nil, Metadata{}, err
	}
	manifestAsset, err := releaseAsset(descriptor, "manifest.json")
	if err != nil {
		return nil, Metadata{}, err
	}
	manifestBytes, err := download(ctx, client, manifestAsset, definitionbundle.MaxManifestSize)
	if err != nil {
		return nil, Metadata{}, err
	}
	manifest, err := definitionbundle.ParseManifest(manifestBytes)
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("release manifest: %w", err)
	}
	if manifest.Version != config.Version {
		return nil, Metadata{}, fmt.Errorf("release manifest version %q does not match pin %q", manifest.Version, config.Version)
	}
	names, err := listCacheNames(root)
	if err != nil {
		return nil, Metadata{}, err
	}
	if err = checkIdentity(root, names, manifest.Version, manifest.Archive.SHA256); err != nil {
		return nil, Metadata{}, err
	}
	if err = manifest.CheckSupport(config.Support.Schema, config.Support.Operations, config.Support.Algorithms); err != nil {
		return nil, Metadata{}, err
	}
	if snapshot, metadata, cacheErr := loadCached(root, names, manifest.Version, config.Support); cacheErr == nil && metadata.Digest == manifest.Archive.SHA256 {
		return snapshot, cacheCurrent(metadata), nil
	}
	archiveAsset, err := releaseAsset(descriptor, manifest.Archive.Name)
	if err != nil {
		return nil, Metadata{}, err
	}
	if archiveAsset.Size != manifest.Archive.Size {
		return nil, Metadata{}, fmt.Errorf("release archive descriptor size %d does not match manifest %d", archiveAsset.Size, manifest.Archive.Size)
	}
	archiveBytes, err := download(ctx, client, archiveAsset, definitionbundle.MaxArchiveSize)
	if err != nil {
		return nil, Metadata{}, err
	}
	files, err := manifest.ReadArchive(archiveBytes, manifest.Archive.SHA256)
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("release archive: %w", err)
	}
	snapshot, err := validateFiles(manifest, files, config.Support)
	if err != nil {
		return nil, Metadata{}, err
	}
	if err := activateSnapshot(root, manifestBytes, archiveBytes, manifest, identity{Version: manifest.Version, Digest: manifest.Archive.SHA256}); err != nil {
		return nil, Metadata{}, err
	}
	return snapshot, Metadata{
		Version: manifest.Version, Digest: manifest.Archive.SHA256,
		Source: "managed-release", CacheStatus: "activated", UpdateStatus: "pinned-new", Updated: true,
	}, nil
}

func loadAutomatic(ctx context.Context, root string, config Config) (*definitions.Snapshot, Metadata, error) {
	snapshot, metadata, err := acquireAutomatic(ctx, root, config)
	if err == nil {
		return snapshot, metadata, nil
	}
	if ctx.Err() != nil {
		return nil, Metadata{}, ctx.Err()
	}
	cached, cacheMetadata, cacheErr := loadAutomaticCache(root, config.Support)
	if cacheErr == nil {
		cacheMetadata.Warning = fmt.Errorf("managed definitions automatic update failed; using validated cached snapshot: %w", err)
		cacheMetadata.UpdateStatus = "cache-after-failure"
		return cached, cacheMetadata, nil
	}
	return nil, Metadata{}, fmt.Errorf("managed definitions automatic update failed (%w); no usable cached snapshot (%v)", err, cacheErr)
}

func acquireAutomatic(ctx context.Context, root string, config Config) (*definitions.Snapshot, Metadata, error) {
	client := newClient(config.Transport)
	releases, err := fetchReleases(ctx, client)
	if err != nil {
		return nil, Metadata{}, err
	}
	for _, candidate := range releases {
		if candidate.Draft || candidate.Prerelease {
			continue
		}
		publishedAt, parseErr := time.Parse(time.RFC3339, candidate.PublishedAt)
		if parseErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, fmt.Errorf("stable release %q has invalid publication time: %w", candidate.TagName, parseErr))
		}
		manifestAsset, candidateErr := releaseAsset(candidate, "manifest.json")
		if candidateErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, candidateErr)
		}
		manifestBytes, candidateErr := download(ctx, client, manifestAsset, definitionbundle.MaxManifestSize)
		if candidateErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, candidateErr)
		}
		manifest, candidateErr := definitionbundle.ParseManifest(manifestBytes)
		if candidateErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, fmt.Errorf("release %q manifest: %w", candidate.TagName, candidateErr))
		}
		if candidateErr = manifest.CheckSupport(config.Support.Schema, config.Support.Operations, config.Support.Algorithms); candidateErr != nil {
			continue
		}
		// Every path past this listing returns, so it is taken at most once.
		names, candidateErr := listCacheNames(root)
		if candidateErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, candidateErr)
		}
		if candidateErr = checkIdentity(root, names, manifest.Version, manifest.Archive.SHA256); candidateErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, candidateErr)
		}
		if snapshot, metadata, cacheErr := loadCached(root, names, manifest.Version, config.Support); cacheErr == nil && metadata.Digest == manifest.Archive.SHA256 {
			return snapshot, cacheCurrent(metadata), nil
		}
		archiveAsset, candidateErr := releaseAsset(candidate, manifest.Archive.Name)
		if candidateErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, candidateErr)
		}
		if archiveAsset.Size != manifest.Archive.Size {
			return nil, Metadata{}, automaticFailure(ctx, fmt.Errorf("release archive descriptor size %d does not match manifest %d", archiveAsset.Size, manifest.Archive.Size))
		}
		archiveBytes, candidateErr := download(ctx, client, archiveAsset, definitionbundle.MaxArchiveSize)
		if candidateErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, candidateErr)
		}
		files, candidateErr := manifest.ReadArchive(archiveBytes, manifest.Archive.SHA256)
		if candidateErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, fmt.Errorf("release archive: %w", candidateErr))
		}
		snapshot, candidateErr := validateFiles(manifest, files, config.Support)
		if candidateErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, candidateErr)
		}
		if candidateErr = activateSnapshot(root, manifestBytes, archiveBytes, manifest, identity{
			Version: manifest.Version, Digest: manifest.Archive.SHA256,
			ReleaseTag: candidate.TagName, PublishedAt: publishedAt.UTC().Format(time.RFC3339),
		}); candidateErr != nil {
			return nil, Metadata{}, automaticFailure(ctx, candidateErr)
		}
		return snapshot, Metadata{
			Version: manifest.Version, Digest: manifest.Archive.SHA256,
			Source: "managed-release", CacheStatus: "activated", UpdateStatus: "update-new", Updated: true,
		}, nil
	}
	return nil, Metadata{}, &acquisitionError{err: fmt.Errorf("no compatible stable managed definition release")}
}

// cacheCurrent marks a validated cached snapshot whose digest matches the
// published manifest, so the archive was not downloaded again.
func cacheCurrent(metadata Metadata) Metadata {
	metadata.Source = "managed-cache"
	metadata.CacheStatus = "current"
	metadata.UpdateStatus = "cache-current"
	metadata.Updated = false
	return metadata
}

func automaticFailure(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	var acquisition *acquisitionError
	if errors.As(err, &acquisition) {
		return err
	}
	return &acquisitionError{err: err}
}

func newClient(transport http.RoundTripper) *http.Client {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &http.Client{
		Timeout:   20 * time.Second,
		Transport: transport,
		CheckRedirect: func(request *http.Request, _ []*http.Request) error {
			if !trustedAssetURL(request.URL) {
				return &redirectError{target: safeURL(request.URL)}
			}
			return nil
		},
	}
}

func fetchRelease(ctx context.Context, client *http.Client, version string) (release, error) {
	endpoint := repositoryAPI + "/releases/tags/" + url.PathEscape(version)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return release{}, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	response, err := client.Do(request)
	if err != nil {
		return release{}, transientError(ctx, requestError(request.URL, err))
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return release{}, transientError(ctx, fmt.Errorf("release descriptor returned HTTP %d", response.StatusCode))
	}
	body, err := readBounded(response.Body, definitionbundle.MaxManifestSize)
	if err != nil {
		return release{}, transientError(ctx, requestError(request.URL, err))
	}
	var result release
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&result); err != nil {
		return release{}, fmt.Errorf("release descriptor JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return release{}, fmt.Errorf("release descriptor JSON must contain one document")
	}
	if result.TagName != version || result.Draft || result.Prerelease {
		return release{}, fmt.Errorf("release descriptor is not the requested stable release %q", version)
	}
	return result, nil
}

func fetchReleases(ctx context.Context, client *http.Client) ([]release, error) {
	// Each stable release keeps the publication time parsed during validation so
	// sorting does not reparse timestamps in the comparator.
	type datedRelease struct {
		release   release
		published time.Time
	}
	releases := make([]datedRelease, 0, releasePageSize)
	tags := map[string]bool{}
	for page := 1; page <= maxReleasePages; page++ {
		endpoint := fmt.Sprintf("%s/releases?per_page=%d&page=%d", repositoryAPI, releasePageSize, page)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Accept", "application/vnd.github+json")
		response, err := client.Do(request)
		if err != nil {
			return nil, transientError(ctx, requestError(request.URL, err))
		}
		body, readErr := readBounded(response.Body, maxReleaseInventorySize)
		closeErr := response.Body.Close()
		if readErr != nil {
			return nil, transientError(ctx, requestError(request.URL, readErr))
		}
		if closeErr != nil {
			return nil, transientError(ctx, requestError(request.URL, closeErr))
		}
		if response.StatusCode != http.StatusOK {
			return nil, transientError(ctx, fmt.Errorf("release inventory returned HTTP %d", response.StatusCode))
		}
		var batch []release
		decoder := json.NewDecoder(bytes.NewReader(body))
		if err = decoder.Decode(&batch); err != nil {
			return nil, automaticFailure(ctx, fmt.Errorf("release inventory JSON: %w", err))
		}
		if err = decoder.Decode(&struct{}{}); err != io.EOF {
			return nil, automaticFailure(ctx, fmt.Errorf("release inventory JSON must contain one document"))
		}
		for _, candidate := range batch {
			if candidate.Draft || candidate.Prerelease {
				continue
			}
			if !versionPattern.MatchString(candidate.TagName) || tags[candidate.TagName] {
				return nil, automaticFailure(ctx, fmt.Errorf("release inventory has invalid or duplicate tag %q", candidate.TagName))
			}
			published, parseErr := time.Parse(time.RFC3339, candidate.PublishedAt)
			if parseErr != nil {
				return nil, automaticFailure(ctx, fmt.Errorf("stable release %q has invalid publication time: %w", candidate.TagName, parseErr))
			}
			tags[candidate.TagName] = true
			releases = append(releases, datedRelease{release: candidate, published: published})
		}
		if len(releases) > releasePageSize*maxReleasePages {
			return nil, automaticFailure(ctx, fmt.Errorf("release inventory exceeds %d releases", releasePageSize*maxReleasePages))
		}
		if len(batch) < releasePageSize {
			break
		}
		if page == maxReleasePages {
			return nil, automaticFailure(ctx, fmt.Errorf("release inventory exceeds %d pages", maxReleasePages))
		}
	}
	slices.SortFunc(releases, func(left, right datedRelease) int {
		if newerRelease(left.published, left.release.TagName, right.published, right.release.TagName) {
			return -1
		}
		if newerRelease(right.published, right.release.TagName, left.published, left.release.TagName) {
			return 1
		}
		return 0
	})
	ordered := make([]release, len(releases))
	for i := range releases {
		ordered[i] = releases[i].release
	}
	return ordered, nil
}

func newerRelease(leftTime time.Time, leftTag string, rightTime time.Time, rightTag string) bool {
	if !leftTime.Equal(rightTime) {
		return leftTime.After(rightTime)
	}
	return leftTag > rightTag
}

func releaseAsset(release release, name string) (asset, error) {
	var selected *asset
	for _, candidate := range release.Assets {
		if candidate.Name == name {
			if selected != nil {
				return asset{}, fmt.Errorf("release asset %q is ambiguous", name)
			}
			if candidate.Size < 0 || !trustedAssetURLString(candidate.BrowserDownloadURL) {
				return asset{}, fmt.Errorf("release asset %q has an unsafe URL or size", name)
			}
			copy := candidate
			selected = &copy
		}
	}
	if selected != nil {
		return *selected, nil
	}
	return asset{}, &acquisitionError{err: fmt.Errorf("release asset %q is missing", name)}
}

func download(ctx context.Context, client *http.Client, asset asset, limit int64) ([]byte, error) {
	if asset.Size > limit {
		return nil, fmt.Errorf("asset %q exceeds byte limit", asset.Name)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return nil, requestError(nil, err)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, transientError(ctx, requestError(request.URL, err))
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, transientError(ctx, fmt.Errorf("asset %q returned HTTP %d", asset.Name, response.StatusCode))
	}
	if response.ContentLength > limit {
		return nil, fmt.Errorf("asset %q exceeds byte limit", asset.Name)
	}
	body, err := readBounded(response.Body, limit)
	if err != nil {
		return nil, transientError(ctx, requestError(request.URL, err))
	}
	if int64(len(body)) != asset.Size {
		return nil, fmt.Errorf("asset %q size %d does not match release descriptor %d", asset.Name, len(body), asset.Size)
	}
	return body, nil
}

func readBounded(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response exceeds byte limit %d", limit)
	}
	return data, nil
}

func transientError(ctx context.Context, err error) error {
	if callerErr := ctx.Err(); callerErr != nil {
		return callerErr
	}
	return &acquisitionError{err: err}
}

type safeRequestError struct {
	target string
	detail string
	err    error
}

func (e *safeRequestError) Error() string {
	if e.target == "" {
		return "managed definitions request failed"
	}
	if e.detail != "" {
		return "managed definitions request to " + e.target + " failed: " + e.detail
	}
	return "managed definitions request to " + e.target + " failed"
}

func (e *safeRequestError) Unwrap() error { return e.err }

func requestError(target *url.URL, err error) error {
	if err == nil {
		return nil
	}
	var redirect *redirectError
	detail := ""
	if errors.As(err, &redirect) {
		detail = redirect.Error()
	}
	return &safeRequestError{target: safeURL(target), detail: detail, err: requestCause(err)}
}

type redirectError struct{ target string }

func (e *redirectError) Error() string {
	return "unsafe managed definitions redirect to " + e.target
}

func requestCause(err error) error {
	for {
		var request *url.Error
		if !errors.As(err, &request) {
			return err
		}
		if request.Err == nil {
			return errors.New("HTTP request failed")
		}
		err = request.Err
	}
}

func safeURL(parsed *url.URL) string {
	if parsed == nil {
		return ""
	}
	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	return strings.ToLower(parsed.Hostname()) + path
}

func trustedAssetURLString(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && trustedAssetURL(parsed)
}

func trustedAssetURL(parsed *url.URL) bool {
	if parsed == nil || parsed.Scheme != "https" || parsed.User != nil {
		return false
	}
	switch strings.ToLower(parsed.Hostname()) {
	case "github.com":
		return strings.HasPrefix(parsed.EscapedPath(), "/BumpyClock/hermes-definitions/releases/download/")
	case "github-releases.githubusercontent.com", "objects.githubusercontent.com", "release-assets.githubusercontent.com":
		return true
	default:
		return false
	}
}

func prepareRoot(root string) (string, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if err = ensureDirectory(absolute); err != nil {
		return "", fmt.Errorf("managed definitions cache: %w", err)
	}
	for _, name := range []string{"locks", "snapshots", "identities"} {
		if err = ensureDirectory(filepath.Join(absolute, name)); err != nil {
			return "", fmt.Errorf("managed definitions cache: %w", err)
		}
	}
	return absolute, nil
}

func ensureDirectory(directory string) error {
	info, err := os.Lstat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(directory, 0o700)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s must be a real directory", directory)
	}
	return nil
}

// validateFiles loads a verified release's definitions in memory, with the
// same selection and loader validation as a directory that WriteDefinitions wrote.
func validateFiles(manifest *definitionbundle.Manifest, files map[string][]byte, support Support) (*definitions.Snapshot, error) {
	if err := manifest.CheckSupport(support.Schema, support.Operations, support.Algorithms); err != nil {
		return nil, err
	}
	suite, err := definitionbundle.ParseSuite(files)
	if err != nil {
		return nil, fmt.Errorf("release conformance envelope: %w", err)
	}
	selected, err := definitionbundle.DefinitionFiles(files)
	if err != nil {
		return nil, err
	}
	// Loader errors keep the prefix they had when the audit performed its own load.
	snapshot, err := definitions.LoadFiles(selected)
	if err == nil {
		err = manifest.AuditRequirements(suite, snapshot)
	}
	if err != nil {
		return nil, fmt.Errorf("release requirements: %w", err)
	}
	return snapshot, nil
}

func snapshotDirectory(root, version, digest string) string {
	return filepath.Join(root, "snapshots", version, digest)
}

// activateSnapshot stores a validated release and records its identity. It
// lists cache names again because a download separates it from the pre-download
// checks, so the alias check stays adjacent to the entries it creates.
func activateSnapshot(root string, manifestBytes, archiveBytes []byte, manifest *definitionbundle.Manifest, value identity) error {
	names, err := listCacheNames(root)
	if err != nil {
		return err
	}
	if err = storeSnapshot(root, names, manifestBytes, archiveBytes, manifest); err != nil {
		return err
	}
	return recordIdentity(root, names, value)
}

func storeSnapshot(root string, names cacheNames, manifestBytes, archiveBytes []byte, manifest *definitionbundle.Manifest) error {
	if err := names.rejectAlias(manifest.Version); err != nil {
		return err
	}
	parent := filepath.Join(root, "snapshots", manifest.Version)
	if err := ensureDirectory(parent); err != nil {
		return err
	}
	target := snapshotDirectory(root, manifest.Version, manifest.Archive.SHA256)
	if _, err := os.Lstat(target); err == nil {
		return validateStoredSnapshot(target, manifest.Archive.SHA256)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	stage, err := os.MkdirTemp(parent, ".staging-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	if err = writeFile(filepath.Join(stage, "manifest.json"), manifestBytes); err == nil {
		err = writeFile(filepath.Join(stage, "bundle.tar"), archiveBytes)
	}
	if err != nil {
		return err
	}
	if err = os.Rename(stage, target); err != nil {
		if errors.Is(err, os.ErrExist) {
			return validateStoredSnapshot(target, manifest.Archive.SHA256)
		}
		return err
	}
	return nil
}

func writeFile(name string, contents []byte) error {
	//nolint:gosec // Destination is a new updater-owned staging or snapshot file.
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	written, err := file.Write(contents)
	if err == nil && written != len(contents) {
		err = io.ErrShortWrite
	}
	return errors.Join(err, file.Close())
}

func validateStoredSnapshot(directory, digest string) error {
	info, err := os.Lstat(directory)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("cached snapshot is not a real directory")
	}
	manifestBytes, err := definitionbundle.ReadFile(filepath.Join(directory, "manifest.json"), definitionbundle.MaxManifestSize)
	if err != nil {
		return err
	}
	manifest, err := definitionbundle.ParseManifest(manifestBytes)
	if err != nil || manifest.Archive.SHA256 != digest {
		return fmt.Errorf("invalid existing cached snapshot")
	}
	return nil
}

// loadCachedVersion validates the cached snapshot for one version with a fresh
// listing of cache names.
func loadCachedVersion(root, version string, support Support) (*definitions.Snapshot, Metadata, error) {
	names, err := listCacheNames(root)
	if err != nil {
		return nil, Metadata{}, err
	}
	return loadCached(root, names, version, support)
}

func loadCached(root string, names cacheNames, version string, support Support) (*definitions.Snapshot, Metadata, error) {
	if err := names.rejectAlias(version); err != nil {
		return nil, Metadata{}, err
	}
	cachedIdentity, hasIdentity, err := readIdentity(root, names, version)
	if err != nil {
		return nil, Metadata{}, err
	}
	directory := filepath.Join(root, "snapshots", version)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, Metadata{}, err
	}
	var loaded []*struct {
		snapshot *definitions.Snapshot
		metadata Metadata
	}
	for _, entry := range entries {
		if !entry.IsDir() || !digestPattern.MatchString(entry.Name()) {
			continue
		}
		if hasIdentity && entry.Name() != cachedIdentity.Digest {
			continue
		}
		snapshot, metadata, loadErr := loadSnapshot(filepath.Join(directory, entry.Name()), support)
		if loadErr == nil && metadata.Version == version && metadata.Digest == entry.Name() {
			loaded = append(loaded, &struct {
				snapshot *definitions.Snapshot
				metadata Metadata
			}{snapshot: snapshot, metadata: metadata})
		}
	}
	if len(loaded) != 1 {
		if len(loaded) == 0 {
			return nil, Metadata{}, fmt.Errorf("no intact cached snapshot for pin %q", version)
		}
		return nil, Metadata{}, fmt.Errorf("ambiguous cached snapshots for pin %q", version)
	}
	return loaded[0].snapshot, loaded[0].metadata, nil
}

func loadAutomaticCache(root string, support Support) (*definitions.Snapshot, Metadata, error) {
	entries, err := os.ReadDir(filepath.Join(root, "identities"))
	if err != nil {
		return nil, Metadata{}, err
	}
	snapshotEntries, err := os.ReadDir(filepath.Join(root, "snapshots"))
	if err != nil {
		return nil, Metadata{}, err
	}
	names := newCacheNames(snapshotEntries, entries)
	type candidate struct {
		known     identity
		published time.Time
	}
	var candidates []candidate
	seen := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		version := strings.TrimSuffix(entry.Name(), ".json")
		key := strings.ToLower(version)
		if previous, exists := seen[key]; exists && previous != version {
			return nil, Metadata{}, fmt.Errorf("case-folded cached release identity collision: %q and %q", previous, version)
		}
		seen[key] = version
		known, exists, identityErr := readIdentity(root, names, version)
		if identityErr != nil || !exists {
			continue
		}
		// Issue36 cache identities predate automatic chronology. They remain
		// usable by their exact pin, but cannot participate in automatic fallback.
		if known.ReleaseTag == "" || known.PublishedAt == "" {
			continue
		}
		published, parseErr := time.Parse(time.RFC3339, known.PublishedAt)
		if parseErr != nil {
			published = time.Time{}
		}
		candidates = append(candidates, candidate{known: known, published: published})
	}
	slices.SortFunc(candidates, func(left, right candidate) int {
		if newerRelease(left.published, left.known.ReleaseTag, right.published, right.known.ReleaseTag) {
			return -1
		}
		if newerRelease(right.published, right.known.ReleaseTag, left.published, left.known.ReleaseTag) {
			return 1
		}
		return 0
	})
	// Validate newest first; a broken newer cache falls through to older ones.
	for _, candidate := range candidates {
		snapshot, metadata, cacheErr := loadCached(root, names, candidate.known.Version, support)
		if cacheErr == nil && metadata.Digest == candidate.known.Digest {
			return snapshot, metadata, nil
		}
	}
	return nil, Metadata{}, fmt.Errorf("no intact compatible cached managed definition snapshot")
}

func loadSnapshot(directory string, support Support) (*definitions.Snapshot, Metadata, error) {
	manifestBytes, err := definitionbundle.ReadFile(filepath.Join(directory, "manifest.json"), definitionbundle.MaxManifestSize)
	if err != nil {
		return nil, Metadata{}, err
	}
	archiveBytes, err := definitionbundle.ReadFile(filepath.Join(directory, "bundle.tar"), definitionbundle.MaxArchiveSize)
	if err != nil {
		return nil, Metadata{}, err
	}
	manifest, err := definitionbundle.ParseManifest(manifestBytes)
	if err != nil {
		return nil, Metadata{}, err
	}
	files, err := manifest.ReadArchive(archiveBytes, manifest.Archive.SHA256)
	if err != nil {
		return nil, Metadata{}, err
	}
	snapshot, err := validateFiles(manifest, files, support)
	if err != nil {
		return nil, Metadata{}, err
	}
	return snapshot, Metadata{
		Version: manifest.Version, Digest: manifest.Archive.SHA256,
		Source: "managed-cache", CacheStatus: "validated", UpdateStatus: "cache-validated",
	}, nil
}

func identityPath(root, version string) string {
	return filepath.Join(root, "identities", version+".json")
}

func readIdentity(root string, names cacheNames, version string) (identity, bool, error) {
	if err := names.rejectAlias(version); err != nil {
		return identity{}, false, err
	}
	data, err := definitionbundle.ReadFile(identityPath(root, version), definitionbundle.MaxManifestSize)
	if errors.Is(err, os.ErrNotExist) {
		return identity{}, false, nil
	}
	if err != nil {
		return identity{}, false, err
	}
	var value identity
	if err = definitionbundle.DecodeJSON(data, &value); err != nil {
		return identity{}, false, err
	}
	if value.Version != version || !digestPattern.MatchString(value.Digest) {
		return identity{}, false, fmt.Errorf("invalid cached release identity for %q", version)
	}
	if value.ReleaseTag != "" && !versionPattern.MatchString(value.ReleaseTag) {
		return identity{}, false, fmt.Errorf("invalid cached release tag for %q", version)
	}
	if value.PublishedAt != "" {
		if _, err = time.Parse(time.RFC3339, value.PublishedAt); err != nil {
			return identity{}, false, fmt.Errorf("invalid cached publication time for %q", version)
		}
	}
	return value, true, nil
}

func checkIdentity(root string, names cacheNames, version, digest string) error {
	known, exists, err := readIdentity(root, names, version)
	if err != nil {
		return err
	}
	if exists && known.Digest != digest {
		return fmt.Errorf("release %q changed known archive identity from %s to %s", version, known.Digest, digest)
	}
	return nil
}

func recordIdentity(root string, names cacheNames, value identity) error {
	if err := names.rejectAlias(value.Version); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	path := identityPath(root, value.Version)
	if old, exists, err := readIdentity(root, names, value.Version); err != nil {
		return err
	} else if exists {
		if old.Digest != value.Digest {
			return fmt.Errorf("release %q changed known archive identity", value.Version)
		}
		return nil
	}
	return writeFile(path, data)
}

// cacheNames is one listing of the version names under snapshots/ and
// identities/, in directory order. One operation may reuse a listing: the cache
// lock excludes other Hermes writers, and this process creates only entries
// named exactly after the checked version, which rejectAlias never reports.
type cacheNames struct {
	snapshots  []string
	identities []string
}

func listCacheNames(root string) (cacheNames, error) {
	snapshots, err := os.ReadDir(filepath.Join(root, "snapshots"))
	if err != nil {
		return cacheNames{}, err
	}
	identities, err := os.ReadDir(filepath.Join(root, "identities"))
	if err != nil {
		return cacheNames{}, err
	}
	return newCacheNames(snapshots, identities), nil
}

func newCacheNames(snapshots, identities []os.DirEntry) cacheNames {
	names := cacheNames{snapshots: make([]string, 0, len(snapshots)), identities: make([]string, 0, len(identities))}
	for _, entry := range snapshots {
		names.snapshots = append(names.snapshots, entry.Name())
	}
	for _, entry := range identities {
		if name := entry.Name(); filepath.Ext(name) == ".json" {
			names.identities = append(names.identities, strings.TrimSuffix(name, ".json"))
		}
	}
	return names
}

// rejectAlias rejects a version that differs only by case from a cached
// snapshot or identity name. Case-insensitive filesystems resolve such names to
// the same entry, which would expose another release's identity or snapshot.
func (names cacheNames) rejectAlias(version string) error {
	for _, location := range [][]string{names.snapshots, names.identities} {
		for _, name := range location {
			if strings.EqualFold(name, version) && name != version {
				return fmt.Errorf("case-folded managed cache version collision: %q and %q", version, name)
			}
		}
	}
	return nil
}
