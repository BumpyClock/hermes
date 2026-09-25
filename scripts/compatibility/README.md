# Released-version compatibility

Compare two immutable engine revisions without switching the working checkout,
accessing live sites, or overwriting historical observations:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 scripts/compatibility/run.py \
  --baseline v1.1.1 \
  --candidate-ref CANDIDATE-COMMIT \
  --output .compatibility-runs/released-candidate \
  --samples 6 --benchtime 1s
```

Use a new output directory for every capture. Go and module dependencies must
already be available locally: the runner disables proxy/checksum downloads and
uses the local toolchain. Archives come from local Git objects, not the network.
Without `--candidate-ref`, it instead records and copies current tracked and
unignored worktree files; use an immutable revision for a final qualification.

The runner records source, fixture, and harness digests, commands, and environment.
It compares 38 public error/format contracts, stable exported API shapes, unchanged
released example compilation, and 30 HTML/Markdown/text observations through both
`Parse` and `ParseHTML`. Each functional capture runs twice to detect
nondeterminism. Literal public-IP URLs and injected HTTP transports avoid DNS and
network requests without modifying production validation.

`intentional-changes.json` binds each reviewed output change to **both** exact
before/after observation digests and the released baseline SHA. New differences,
unused allowances, changed stable APIs, or unexpected additions fail the run.
Do not regenerate this file to make a failure disappear; first diagnose the
specific observable change and establish whether it is an intended repair.

Nine generic extraction benchmarks use identical fixture bytes, reusable
clients, and alternating baseline/candidate order. At least five samples per
workload are required. Compare medians together with ranges and allocations;
shared-machine timing is not proof of a causal performance regression or speedup.
Generic fixture URLs deliberately avoid compiled-site selection. They must not
be described as configured-YAML performance measurements.

Generated `.compatibility-runs/` data is ignored by Git. Preserve the JSON reports,
raw observations, benchmark logs, and provenance for a reviewed capture; extracted
source/build trees and binaries are disposable scratch.

Recheck existing functional evidence without recapturing:

```sh
python3 scripts/compatibility/run.py \
  --output .compatibility-runs/released-candidate --verify-only
make test-compatibility
```

Use `--functional-only` on a new capture to replay API, consumer, and output
contracts without collecting another set of timings.

`make verify` includes the comparator's offline unit tests, not the full
release comparison or benchmarks. The latter need an explicit baseline,
candidate, and evidence directory.

## Configured-site extraction

Compare the old compiled NYTimes and Ars Technica rules with an explicitly
loaded external snapshot:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 scripts/compatibility/run_configured.py \
  --candidate-ref e4560ca2774c359ff66af51d4d5f8408c0eedbb9 \
  --definitions /absolute/path/to/definitions-fb626db \
  --definitions-source fb626db414b875dcddad110174303d1e1bd34e84 \
  --output .compatibility-runs/configured-candidate \
  --samples 6 --benchtime 1s
```

The definitions input is a whole immutable checkout/archive, containing the
`definitions/` directory. Exact rule and fixture digests are recorded. Loading
and parser construction occur outside the timers. Both versions use the
internal parser's extraction entry point and the original site hostnames; the
runner verifies that a compiled or external rule actually matched, rather than
silently measuring generic fallback.

An identical build-only overlay replaces one checked DNS resolver call in both
immutable snapshots. Its original and substituted source digests are recorded.
This avoids DNS latency without editing either checkout or exposing a public
validation bypass. These are extraction-only measurements, not public network,
startup-loading, or full-corpus performance measurements.

The six workloads (two historical fixtures by three formats) report exact output
differences, repeated samples, timing, and allocations. The tool does not assert
performance equivalence. Review meaningful differences before accepting a
candidate, and retain earlier reports when investigating an optimization.
Pass `--reference .compatibility-runs/PRIOR-CONFIGURED-CAPTURE` to require exact
equality to the prior candidate's complete configured result objects as well.

## Comparing behavior-preserving optimizations

Use pairwise mode when both sides must produce exactly the same results. This
does not reuse or expand the released-version intentional-change allowlist.
For an uncommitted candidate:

```sh
BASE=acf278d0abc1795a4c872b6cdcd77d204e071f13

PYTHONDONTWRITEBYTECODE=1 python3 scripts/compatibility/run.py \
  --baseline "$BASE" --candidate-worktree --comparison-mode pairwise \
  --output .compatibility-runs/optimization-generic \
  --samples 6 --benchtime 1s

PYTHONDONTWRITEBYTECODE=1 python3 scripts/compatibility/run_configured.py \
  --baseline "$BASE" --candidate-worktree \
  --baseline-mode definitions --require-equivalent \
  --definitions /absolute/path/to/definitions-fb626db \
  --definitions-source fb626db414b875dcddad110174303d1e1bd34e84 \
  --output .compatibility-runs/optimization-configured \
  --samples 6 --benchtime 1s
```

Both configured versions load the same snapshot outside the timers. Add
`--functional-only` for an intermediate correctness check without timings.
Pairwise comparison checks the full result objects, API shapes, workload
identities, and repeated observations, not just selected content substrings.

The worktree option captures tracked and unignored regular files into an
explicit source archive. It verifies copied bytes and concurrent edits before
acceptance, then checks runtime source drift throughout execution. A drift
failure stops the run; it never silently recaptures a different candidate.
Preserve `candidate.tar` (or a losslessly compressed copy) and its recorded
digest: a commit ID alone cannot reproduce uncommitted source.
