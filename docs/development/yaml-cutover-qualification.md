# External-only YAML cutover qualification

This records local migration qualification, not authorization to publish an
engine or definitions release. No compiled site definitions remain. Consumers
must explicitly load YAML to retain site-specific extraction; unconfigured
clients are generic-only.

## Immutable inputs

| Input | Identity |
| --- | --- |
| Released engine baseline | `v1.1.1`, commit `0bdaeb9567d5f018947252728303fad54f0b0484` |
| Complete candidate runtime | `e4560ca2774c359ff66af51d4d5f8408c0eedbb9` |
| Definitions source | `BumpyClock/hermes-definitions`, commit `fb626db414b875dcddad110174303d1e1bd34e84` |
| Prepared source SHA-256 | `fc1dad2227f799d833584d6edf515a0c3a4fb7ebfe7e003027441b8c6b7baff6` |
| Prepared candidate build ID | `75e5e4120573e9a95b911af8d4a3fdf64f76958aefefec3d27090c016052a84a` |
| Prepared executable SHA-256 | `5217924215bc8b49355032a3caf4758b3134a6a37f423c033d55fca764170bd2` |
| Conformance report SHA-256 | `85e18e534b6efdd0be5f10b7c76758f755f8ccdb35403bc9bf2eec51a82adf62` |
| Full-corpus manifest SHA-256 | `6c075acd18fc6fe4220a7b4e17d80bf2d6e83af8c0e5b62fb5d69d174644316a` |
| Full-corpus archive SHA-256 | `1db1084e880697ed55975d832bd0f754a80e2fed17ef1599b69f514ac054151b` |

The runtime identity includes ordered transforms, registry removal, bundle
validation, and managed loading. Later test/tooling/documentation commits do not
change that production source. Candidate preparation uses an immutable engine
revision and records its offline resolver substitution explicitly; it does not
silently build from a dirty checkout.

## Corpus and compatibility

The definitions-side qualification reconciles all **125 definitions** and all
**161 frozen canonical/alias host strings**. Its 810 actual-matcher checks include
case, `www`, final-dot variants, and declared YAML patterns. This audit found and
repaired a missing Wikipedia apex host; the language-subdomain wildcard alone
did not cover `wikipedia.org`.

The 229 canonical cases pass in HTML, Markdown, and text: **687 case/format
checks**, with zero HTTP attempts and 1,374 in-memory DNS lookups. Repeated
candidate and bundle preparation reproduced identical executable, manifest,
archive, and conformance-report identities.

Current-layout evidence has four disjoint dispositions:

| Disposition | Definitions |
| --- | ---: |
| Fresh validated | 44 |
| Historical fresh-capture-only evidence, retained under its original label | 16 |
| Repaired | 27 |
| Documented current-layout exceptions | 38 |

These sum to 125; they are not 125 claims of fresh live-site success. The full
candidate remains `scope: pilot`, `production_eligible: false`. Complete
production approval is still enforced separately and has not been fabricated or
weakened to accept synthetic fixtures or inaccessible-site exceptions.
Historical live captures were not re-fetched or replayed against the final
engine. The final replay covers the retained canonical corpus; earlier live
observations retain their original engine and evidence provenance.

The frozen 36-observation site baseline has 18 unchanged and 18 changed rows.
Reviewed changes preserve Deadspin's previously omitted YouTube link, separate
NPR Markdown image/caption text (including word count 61 to 62), and preserve
plain-text block boundaries. The comparison's nonzero exit and exact field
differences are retained; original goldens were not rewritten to hide changes.

See the definitions repository's `docs/full-corpus-qualification.md` and
`evidence/qualification/final/` for the reviewed differences, source provenance,
active evidence verifier, Python commands, and reproduction procedure.
The final handoff is committed as
`44bebe53d694dafafa7060b6692e110c73e13366`. Its 93 Python tests pass with the
explicit current, historical, and host-audit inputs; no tests were skipped.
The earlier `bc6d186` evidence remains byte-for-byte preserved under
`evidence/qualification/history/bc6d186/`. Requalification after the single-pass
sanitizer optimization confirms exact equality of all 36 historical observation
rows to that prior candidate, not merely equal pass or changed-row counts.

## Engine gates

The local environment is Go 1.27.1 on macOS/arm64 with golangci-lint 2.13.2
(built with Go 1.27.0).

| Check | Result |
| --- | --- |
| `make verify` after runtime commits | Passed: zero lint issues, race/coverage suite, CLI build, and 17 offline release-tool tests |
| Exact content node, depth, and cumulative-work limits | Below/at-limit accepted; above-limit rejected with contextual resource errors |
| Fresh race tests for definitions, updater locks, and CLI | Passed |
| Canonical public `Parse`/`ParseHTML` pilots against the immutable definitions directory | Passed with race detection |
| Full 125-definition local, exact-pin, automatic, and cached startup modes | Passed with race detection; all 229 case host bindings retained |
| Acquisition failure with validated cache | Warning and identical snapshot identity |
| Acquisition failure without cache | Fatal, with no success snapshot |
| Windows/amd64 and Linux/amd64 public integration test binaries | Cross-compiled successfully |

Cross-compilation is not native Windows or Linux runtime lock verification.
Managed acquisition tests use injected transports, not a newly published release.
The prepared candidate's offline network boundary does not weaken production
URL validation.

For full engine acceptance:

```sh
make verify
go test -race ./internal/definitions ./internal/definitionupdate ./cmd/hermes -count=1
```

Prepare an immutable definitions checkout with `git archive` at the source
revision above, then run the opt-in public integration checks:

```sh
HERMES_DEFINITIONS_PILOT=/absolute/path/to/definitions-fb626db \
  go test -race . -run '^TestCanonical' -count=1 -v
```

Using the full-corpus artifact produced by the definitions-side procedure:

```sh
HERMES_DEFINITION_CANDIDATE=/absolute/path/to/full-corpus-e4560ca \
HERMES_DEFINITION_MANIFEST_SHA256=6c075acd18fc6fe4220a7b4e17d80bf2d6e83af8c0e5b62fb5d69d174644316a \
  go test -race . -run '^TestManagedDefinitionsCandidate$' -count=1 -v
```

The opt-in gate requires an exact manifest digest and validates archive bytes;
it does not discover a latest snapshot. Normal tests skip these external-input
gates when their explicit inputs are absent.

## Released-version comparison

The offline compatibility runner compared immutable v1.1.1 and `bc6d186`
sources with the same Go toolchain, fixture bytes, and harness:

| Check | Result |
| --- | --- |
| Deterministic public error/format contracts | 38 of 38 unchanged; two captures per version agree |
| Exported API shapes | All 21 released declarations unchanged; 11 additive definition APIs |
| Released consumers | All four unchanged example packages compile against both versions |
| Generic fixture observations through both public entry points | 16 unchanged; 14 reviewed changes across 52 leaf fields; no unexpected differences |

The final functional replay at `e4560ca` passes the same contracts, API checks,
consumer compilation, and unchanged exact fixture allowlist without
nondeterminism.

The reviewed fixture changes are text block separation/whitespace, literal
Markdown encoding, and empty-link/link-whitespace serialization. Result helpers,
excerpts, and word counts reflect changed content. HTML content and author/date
values in this comparison are unchanged. The exact before/after digests are
allowlisted in `scripts/compatibility/intentional-changes.json`; changed or stale
allowances fail rather than silently updating the baseline.

Nine identical generic workloads (three fixtures and three formats) ran six
one-second samples per version in alternating order, with `GOMAXPROCS=1` and
`-test.cpu=1`. The immutable replay measured **1.75% to 4.48% higher median
latency**. All ranges overlap; an earlier run of identical runtime bytes ranged
from 5.02% faster to 0.13% slower. Shared-machine measurements therefore do not
establish a stable causal latency regression or speedup. They are not evidence
of identical performance.

Median bytes per operation changed from -0.20% to +2.26%; allocation counts
changed from -0.17% to +0.31%. The reproducible maximum byte increase is NYTimes
generic text output, consistent with the extra block-boundary conversion work.
These are generic measurements, not configured-site, loading, or network timings.

Reproduce with:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 scripts/compatibility/run.py \
  --baseline v1.1.1 \
  --candidate-ref bc6d1866871a50195fb5516eb5188527f3eca59d \
  --output .compatibility-runs/released-candidate \
  --samples 6 --benchtime 1s
```

See the [compatibility runner](../../scripts/compatibility/README.md) for capture
provenance, exact-difference acceptance, and offline test commands.

### Configured-site performance

The final configured comparison uses v1.1.1 compiled rules versus the entire
explicit `fb626db` YAML snapshot on runtime `e4560ca`. Both sides use identical
historical NYTimes (63,114 bytes) and Ars Technica (32,782 bytes) fixtures,
actual site hostnames, and the same build-only DNS control. Six alternating
one-second samples per workload exclude loading and parser construction.
All six complete configured output objects exactly match the prior `bc6d186`
candidate.

| Workload | Median latency change | Bytes/op change | Allocations/op change |
| --- | ---: | ---: | ---: |
| NYTimes HTML | +0.29% | +1.60% | -0.69% |
| NYTimes Markdown | +5.39% | +6.37% | +11.57% |
| NYTimes text | +9.03% | +9.84% | +23.46% |
| Ars HTML | +0.82% | -0.46% | +4.84% |
| Ars Markdown | +5.31% | +5.32% | +9.54% |
| Ars text | +6.52% | +7.10% | +14.28% |

Profiling found that configured HTML was sanitized twice. The final runtime
removes that redundant pass while retaining the old boundary-whitespace result;
the broader historical comparison caught the whitespace regression and a
failing-then-passing test now covers it. Final HTML timing ranges overlap the
baseline.

The remaining Markdown/text timing ranges do not overlap the baseline:
**there is a measured 5.31-9.03% latency cost**, not performance equivalence.
Paired allocation profiles attribute the dominant extra NYTimes text
allocations to the required pre-conversion sanitizer, with block-boundary text
handling also contributing. The sanitizer policy was not weakened to recover
the old cost. These are two-site extraction measurements, not all-site,
network, loading, concurrency, or production-throughput claims.

Reproduction and exact-output reference comparison are documented in the
compatibility runner. Compact raw evidence remains under
`.compatibility-runs/configured-v1.1.1-e4560ca-fb626db/`; the final generic
functional evidence is under
`.compatibility-runs/released-v1.1.1-pinned-e4560ca/`.

## Deployment boundary

Follow the [consumer migration guide](../guides/yaml-migration.md). Public module
path, existing API signatures, and result JSON shape remain stable, but default
site selection deliberately changes. Loading an external snapshot is required
for migrated site behavior. Literal-output and metadata repairs are intentional
differences, not a promise of byte-for-byte legacy output.

No release, tag, push, deployment, module-major migration, or complete-production
approval is implied by this local qualification.
