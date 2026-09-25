# Independent definition candidate conformance

The [full cutover qualification](yaml-cutover-qualification.md) records the
immutable 125-definition candidate, consumer-mode checks, and evidence limits.

`cmd/definition-conformance` is an **offline development test runner**, not a
production Hermes binary or a release asset. It validates a caller-selected
manifest and archive with `internal/definitionbundle`, invokes the real
`hermes.LoadDefinitions` / `WithDefinitions` / `ParseHTML` path, and emits a JSON
report for the exact engine/bundle pair. Ordinary builds refuse to run it without
the explicit candidate preparation boundary and build provenance.

The canonical producer and site cases live in `BumpyClock/hermes-definitions`:

- `scripts/bundle/prepare_engine.py`: build from an immutable local engine
  revision, optionally adding only the explicitly supplied conformance-owned
  source trees.
- `scripts/bundle/bundle.py`: prepare/check a deterministic candidate archive
  with an explicitly selected executable and receipt.
- `docs/bundle-protocol.md`: protocol 1 encoding, capability and complete-release
  gate contracts.
- `docs/candidate-bundles.md`: preparation commands and offline CI recipe.
- `docs/current-candidate-conformance.md`: current BBC+NYTimes acceptance using
  separately pinned runtime34 and conformance33 source overlays.

No command discovers a release, downloads a candidate, publishes assets, updates
a cache, activates production definitions, or changes Hermes's source-only
stable-v1 release convention.

## Engine-owned tests and prepared inputs

`internal/definitionbundle/testdata/pinned` is a versioned **synthetic pilot**,
not a production definition set or canonical site fixture. Its recorded pins are:

| Input | SHA-256 |
| --- | --- |
| `manifest.json` | `bcab7a561cd1633e75104133361ed8ae855e1aaf9d3d7ab31bedd1702ef633c2` |
| `bundle.tar` | `7e3ac7726f3a07bc60d6795798cd781d3b7dae90e4a7f8515ecd841131a4b3b2` |

The fixture was explicitly prepared using the definitions-side producer at base
revision `8a2c997db78011cd41be6cae0fe4334b5a6f4119` plus its candidate tooling
overlay. Its `source.state: overlay` describes synthetic inputs, not pristine
production content. No engine identity is fabricated in this fixture.

Preparation is separate from ordinary tests, and requires an explicit local
producer checkout:

```sh
python3 internal/definitionbundle/testdata/prepare.py \
  --producer-root ../hermes-definitions \
  --producer-revision 8a2c997db78011cd41be6cae0fe4334b5a6f4119
```

Review both fixture bytes and changed pins when intentionally regenerating.
Tests never regenerate them implicitly.

```sh
mkdir -p bin/conformance-test-work
TMPDIR="$PWD/bin/conformance-test-work" GOTMPDIR="$PWD/bin/conformance-test-work" \
  GOPROXY=off GOSUMDB=off \
  go test ./internal/definitionbundle/... ./cmd/definition-conformance -count=1
```

These tests check bounded envelopes/archives, the actual loader, capabilities,
case diagnostics, exact complete-cohort approval, resolver behavior, and rejection
of an ordinary, unprepared runner. `TestOfflineBoundarySecurity` deliberately
skips in an ordinary build rather than letting public URL validation perform
DNS. Run it in the explicitly prepared source snapshot as described below.

## Offline network boundary

Public `ParseHTML` validates URLs, including DNS, even for supplied HTML. A
deny-HTTP transport alone is therefore insufficient.

Preparation snapshots the exact pinned source and records its complete file
content digest. It replaces **one checked resolver construction**, only in that
snapshot, with `internal/definitionbundle/offline.Resolver`, and adds a small
snapshot-only installation shim. The production `internal/validation/url.go`
file in the working checkout is never edited. Two non-build agent-skill symlinks
are explicitly excluded and recorded, never followed.

The injected resolver:

- Answers only the validated suite's fixture hostnames, entirely in memory.
- Preserves literal IPs, allowing existing private-address/localhost denial to
  run unchanged.
- Rejects unplanned hostnames and honors context cancellation.
- Does not use system DNS, open a socket, or fall back to a real resolver.

The runner also denies both its explicit HTTP transport and the process default
transport. Any HTTP attempt fails the run even if a parser were to suppress that
error. The report records attempted HTTP operations and in-memory lookup counts.
The prepared-snapshot security test invokes public `ParseHTML`, accepts a fixture
hostname, and verifies `ErrInvalidURL` for private IPs and an unplanned hostname.

This mechanism is a **test boundary**, not permission to weaken production SSRF
validation or add a public bypass option.

## Consumer integration

The manifest is strict JSON. The archive is uncompressed, bounded USTAR with
sorted regular entries, normalized metadata and no links or extensions.
Wire keys are exact before typed decoding; case-folded aliases cannot replace
capability lists. All ASCII payload paths must be unique after case folding,
and validated YAML is written with exclusive creation, never truncation.
`ParseManifest`, `CheckSupport`, `ReadArchive`, `ParseSuite`,
`AuditRequirements`, and `CheckCoverage` fail closed. Requirements auditing uses
the real loader and matcher; it does not implement a second extraction engine.
Only verified YAML is written into a new caller-selected directory.

The real loader records each capability a definition uses while it validates
that definition, and `AuditRequirements` checks the loaded snapshot's recorded
set. There is no separate YAML walker to keep in sync with the validator.
Operation keys require `transform.<key>` and conditions require `condition.<key>`.
The selected engine's authoritative capability set must contain every declared
requirement; missing declarations also fail. Schema 1 alone does not authorize
transforms on an older engine. Named algorithms are checked independently
against the selected engine's reported algorithm IDs.

Cases may select HTML, Markdown and text formats explicitly. Format-scoped
`contains` assertions preserve canonical NYTimes URL checks in HTML/Markdown
without imposing those URL fragments on plain text. Existing `exactly_once` and
`exclude` assertions apply unchanged. Every selected format needs a positive
assertion, and each report row identifies the format.

The real matcher receives `URL.Hostname()` without pre-normalization, matching
the public parser. A repeated trailing dot cannot certify a generic-fallback
result as a site match.

## Explicit current-runtime provenance

`scripts/bundle/source_overlay.py` in the definitions repository captures a
caller-selected sorted file list into a manifest of exact sizes and SHA-256s.
`prepare_engine.py --source-root ... --source-overlay MANIFEST=SHA256` accepts
multiple separately named overlays, verifies their pins/base revision/content,
rejects path collisions and overlapping files, and copies only verified bytes
over the immutable base snapshot. Two-pass reads detect source changes during
capture/verification and require an explicit retry; they do not recapture
silently. The runtime34 and conformance33 overlays are separately recorded.

The original conformance-only `--overlay-root` remains available for baseline
negative controls, but cannot be combined with explicit source-overlay manifests.
No mode implicitly copies an entire dirty working tree. The prepared source
digest, overlay manifests, builder-source digests, Go target/version, build ID
and executable digest identify the actual candidate rather than a clean-ref
claim. Runtime source and canonical fixtures are not edited by preparation.

Issue 36 must preserve explicit version/digest identity, capability validation,
bounded archive validation, and the distinction between incomplete pilots and
approved complete sets. Candidate executable paths and build commands are trusted
caller inputs, never fields read from a bundle. Acquisition, cache activation,
automatic stable selection, rollback and publication remain outside this runner.
