package hermes

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/BumpyClock/hermes/internal/definitionupdate"
)

// ManagedDefinitionsOptions selects an exact or automatically discovered public
// release and its updater-owned cache. Loading never performs article requests.
type ManagedDefinitionsOptions struct {
	// Version is the exact stable hermes-definitions release tag to load.
	Version string
	// Automatic selects the newest compatible stable release by publication
	// chronology. It cannot be combined with Version.
	Automatic bool
	// CacheDirectory is updater-owned storage. An empty value uses the OS cache.
	CacheDirectory string
	// Transport is an optional acquisition transport for deterministic tests.
	// URLs and redirect destinations remain restricted to the official publisher.
	Transport http.RoundTripper
}

// ManagedDefinitions is a successful immutable snapshot and its startup
// provenance. Snapshot is non-nil whenever LoadManagedDefinitions returns nil error.
type ManagedDefinitions struct {
	Snapshot     *Definitions
	Version      string
	Digest       string
	Source       string
	CacheStatus  string
	UpdateStatus string
	Updated      bool
	Warning      error
}

// LoadManagedDefinitions downloads or revalidates one compatible public release
// before client construction. A failed acquisition can use only an intact
// compatible cache (matching Version when pinned), reported through Warning.
func LoadManagedDefinitions(ctx context.Context, options ManagedDefinitionsOptions) (*ManagedDefinitions, error) {
	cacheDirectory := options.CacheDirectory
	if cacheDirectory == "" {
		userCache, err := os.UserCacheDir()
		if err != nil {
			return nil, fmt.Errorf("managed definitions cache directory: %w", err)
		}
		cacheDirectory = filepath.Join(userCache, "hermes", "definitions")
	}
	support := DefinitionCapabilities()
	operations := append([]string(nil), support.Capabilities...)
	algorithms := append([]string(nil), support.Algorithms...)
	slices.Sort(operations)
	slices.Sort(algorithms)
	snapshot, metadata, err := definitionupdate.Load(ctx, definitionupdate.Config{
		Version: options.Version, Automatic: options.Automatic, CacheDirectory: cacheDirectory, Transport: options.Transport,
		Support: definitionupdate.Support{Schema: support.Schema, Operations: operations, Algorithms: algorithms},
	})
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, fmt.Errorf("managed definitions loader returned an empty snapshot")
	}
	return &ManagedDefinitions{
		Snapshot: &Definitions{snapshot: snapshot},
		Version:  metadata.Version, Digest: metadata.Digest, Source: metadata.Source,
		CacheStatus: metadata.CacheStatus, UpdateStatus: metadata.UpdateStatus,
		Updated: metadata.Updated, Warning: metadata.Warning,
	}, nil
}
