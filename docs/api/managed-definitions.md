# Managed definition releases

Managed definitions are an opt-in, startup-only source for an exact public
release or the newest compatible stable release from
`BumpyClock/hermes-definitions`. They do not change
`New(opts ...Option) *Client`, create background polling, or add definition
I/O to article parsing. Clients without an explicit local or managed snapshot
remain generic-only.

```go
managed, err := hermes.LoadManagedDefinitions(ctx, hermes.ManagedDefinitionsOptions{
    Version: approvedReleaseTag,
})
if err != nil {
    return err
}
if managed.Warning != nil {
    log.Printf("using cached definitions %s (%s): %v",
        managed.Version, managed.Digest, managed.Warning)
}
client := hermes.New(hermes.WithDefinitions(managed.Snapshot))
```

`approvedReleaseTag` must name an actually published, approved release; development
candidate names in conformance reports are not downloadable releases.
When provided, `Version` is an exact release tag, not a range. The returned snapshot is immutable and non-nil on success. Its
`Version`, archive `Digest`, `Source`, `CacheStatus`, `Updated`, and optional
`Warning` identify the loaded content and whether startup activated a fresh
download or used a validated cache. A pinned load reports `UpdateStatus`
`pinned-new` after activating a download, `cache-current` when the published
manifest digest matches the validated cached snapshot (the archive is not
downloaded again), or `cache-after-failure`.

Set `Automatic: true` instead of `Version` to check the bounded public release
inventory at startup. Hermes rejects draft and prerelease entries, validates
each candidate's manifest and current engine capabilities, and selects the
newest compatible release by GitHub `published_at` timestamp; equal timestamps
break ties by descending release tag. This is not a SemVer comparison and does
not use a blind "latest" endpoint. `UpdateStatus` reports `update-new`,
`cache-current`, or `cache-after-failure`; a warning means that no successful
update check occurred.

## Trust, validation, and caching

The loader queries the official GitHub publisher over HTTPS, selects the
explicit release descriptor, then requires its `manifest.json` and exactly the
archive asset named by that manifest. It bounds requests, redirects, response
sizes, archive entries, and extracted bytes. The existing bundle validator
rejects unsafe archives, malformed strict JSON, case-folded paths, invalid YAML,
domain collisions, and unsupported engine operations or named algorithms before
activation.

Redirect diagnostics identify only a destination host and path. They redact
query parameters, user information, and fragments so signed asset URLs do not
reach returned errors, cache warnings, or CLI stderr.

Checksums establish the selected archive's byte integrity; they do not protect
against a compromised trusted publisher. Publisher signatures are not supported.
Automatic discovery uses the same publisher and validation boundaries as exact
pins.

Snapshots are content-identified and stored under an updater-owned cache.
Staging is validated before atomic activation. Each startup revalidates the
cached manifest/archive bytes, engine capabilities, and definitions of any
snapshot it returns. A release tag
whose known archive digest changes is rejected. The loader does not alter local
definition directories, merge local and managed sources, remove cached releases,
or replace existing clients.

Managed cache coordination uses advisory process locks on macOS and supported
BSD/Linux targets, and `LockFileEx` on Windows. Lock waits honor the caller
context, and operating-system handle closure releases an abandoned owner lock.
Other targets return an explicit managed-loading error rather than proceeding
without cross-process coordination.

The Windows implementation is cross-compiled as part of portability checks.
Those checks do not substitute for a native Windows runtime lock test when
changing its locking behavior.

If acquisition has a transient failure (including an outage or rate limit), the
loader may return only an intact, compatible cached snapshot for the same pin.
It then sets `Warning`; no usable exact cache is fatal. Cancellation remains
fatal and is returned as the context error rather than a cache success.

Automatic mode can similarly retain the newest intact compatible cached
snapshot when inventory discovery or a candidate download/validation fails.
Cache fallback validates cached snapshots only after acquisition fails.
Automatic fallback checks cached releases newest first and skips corrupt or
incompatible entries.
Pinned mode is the deliberate rollback control: restart with an older exact
`Version`; it never substitutes a newer or different release.

## CLI

```sh
hermes parse \
  --managed-definitions-auto \
  --definitions-cache "$HOME/Library/Caches/hermes/definitions" \
  https://example.com/article
```

The CLI loads managed definitions once before workers start. A warning is sent
to stderr and JSON output remains unchanged. Fatal loading errors occur before
article requests. `--definitions` (offline local source) and
`--managed-definitions` are mutually exclusive; `--definitions-cache` requires
the managed source. `--managed-definitions` and `--managed-definitions-auto`
are mutually exclusive.
