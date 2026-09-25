# Refactor comparison

The comparison tool checks two local worktrees against the same HTML fixtures. It does not fetch article URLs.

## Run the comparison

Use Python 3.11 or later and the Go version from `go.mod`.
Keep both worktrees unchanged while the tool executes.

```sh
python3 scripts/compare_contract.py \
  --base ../hermes \
  --candidate . \
  --output /tmp/hermes-contract-comparison
```

The output directory must not exist. The tool preserves earlier captures.

The cleanup base is commit `0bdaeb9567d5f018947252728303fad54f0b0484`.
If the original worktree has changed, create a detached baseline:

```sh
git worktree add --detach ../hermes-baseline 0bdaeb9567d5f018947252728303fad54f0b0484
```

Use `--base ../hermes-baseline` for that worktree.

## Compared contracts

| Check | Coverage |
| --- | --- |
| Public declarations | Exact `go doc -all .` output, including public API documentation |
| Fixture results | Every public `Result` field, `FormatMarkdown`, and result helper methods |
| Extraction routes | Original fixture domain and a generic-only numeric host |
| Input routes | `ParseHTML` and `Parse` with an injected HTTP transport |
| Output formats | HTML, Markdown, and text |
| Error contracts | Codes, messages, operation, URL, unwrap chain, and error helper methods |
| Boundary cases | Invalid URLs, private IP denial, format aliases, HTTP status, MIME type, empty HTML, and canceled contexts |
| Retry failures | Transport failure and body-read failure, including four-attempt exhaustion |

The repository has 170 HTML fixtures. The full fixture pass produces 2,040 observations per worktree.
The unmodified-build contract pass produces 38 observations per worktree.
The comparator does not ignore result fields or normalize returned strings.

## Controlled fixture environment

The contract pass uses the unmodified library. Its URLs contain numeric IP addresses, so it needs no DNS lookup.

The fixture pass preserves site hostnames so custom extractors remain active.
A Go build overlay substitutes a fixed public IP for DNS lookup. HTTP responses come from the fixture transport.
The overlay replaces no other source file and does not change the source worktree.

Attribute traversal is deterministic. These functions iterate `node.Attr` in source order instead of the map from `dom.GetAttrs`:

- `ConvertNodeTo` in `internal/utils/dom/convert.go` emits attributes in source order.
- `ConvertLazyLoadedImages` in `internal/resource/dom.go` visits image attributes in source order. When several attributes match, the last one sets `src` or `srcset`.
- `CleanAttributes` in `internal/utils/dom/clean.go` removes attributes in place, so kept attributes stay in source order.

Both fixture builds therefore use production traversal order, and the tool has no `--map-order` option.
Earlier revisions iterated these maps, so their fixture output can differ between two executions of the same revision.
A comparison between such a revision and a later one also reports the attribute and image differences that the source-order change introduced.
To compare two earlier revisions under the former sorted-key control, run `scripts/compare_contract.py` from one of those revisions.
Relative dates can also depend on the current clock. Treat unexplained differences as unresolved, not as acceptable output changes.

## Evidence

The output directory contains the fixture hashes, revision IDs, diff summaries, API text, JSONL observations, build logs, and comparison report.
The process returns a nonzero status for a build failure, missing case, API change, or observation difference.

Run the comparator tests separately:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts -p 'test_compare_contract.py' -v
```

Run `make verify` for lint, race tests, CLI build, and release-tool checks.
Run `npm test --prefix benchmark` for the benchmark runner contract.

## Cleanup acceptance

The `refactor/simplify-core` branch preserves all public API source files and module dependencies.
The controlled comparison matched 2,040 fixture observations and 38 unmodified-build contract observations.
The reverse-order Washington Post comparison matched 12 fixture observations and the same 38 contracts.

The raw comparison differed in eight observations. Two independent baseline captures differed in eleven observations.
Those differences came from existing attribute order and image selection. This cleanup does not change those policies.
These runs used the former `--map-order` option with its `sorted`, `reverse`, and `raw` modes.

The acceptance checks passed:

- `make verify`: lint, race tests with coverage, CLI build, and 17 release-tool tests.
- `npm test --prefix benchmark`: five tests.
- `python3 -m unittest discover -s scripts -p 'test_compare_contract.py' -v`: three comparator tests.

The cleanup removes unused implementations, per-field goroutines, duplicate title algorithms, and duplicate result assembly.
Typed selectors replace ambiguous metadata and content values. The benchmark runner shares execution and report code.
The initial CLI and API example changes preserved empty-batch JSON and existing format-case behavior.
The review follow-up intentionally fixes the API example's case variants: `JSON`, `Html`, `MarkDown`, and `TEXT` now use the same parser and response media type as their lowercase forms.
It also preserves client-configured resource headers with case-insensitive request overrides and reuses the default HTTP client for direct internal parser calls.
These corrections have focused HTTP tests; the historical comparison results above describe the initial cleanup.

The response buffer pool remains. Five repeated body-read samples showed more allocations and higher latency with `io.ReadAll`.
DOM tag replacement and distinct custom/generic cleanup policies also remain because their proposed replacements can change output.

### Parser benchmark results

The benchmark used Go 1.27.1 on darwin/arm64, with an Apple M5 Max.
Both builds used a DNS-only overlay and the same fixtures. Other comparison and test processes had completed.
Each fixture had five samples at one CPU and five samples at eight CPUs, with a 200 ms benchmark duration.

The command used `BenchmarkParseMultipleFixtures`, `-run '^$'`, `-benchmem`, `-benchtime=200ms`, `-count=5`, and `-cpu=1` or `-cpu=8`.
The `-overlay` argument selected the generated DNS substitution. These measurements exclude DNS latency.

| Fixture | Median latency change, 1 CPU | Median latency change, 8 CPUs | Allocation reduction per parse |
| --- | ---: | ---: | ---: |
| New York Times | -1.19% | +0.61% | 16–19 |
| Washington Post | -0.35% | +1.14% | 17–18 |
| CNN | +0.12% | +3.58% | 16–19 |
| Medium | +0.47% | -1.41% | 16–18 |
| Ars Technica | -1.83% | +1.50% | 16–17 |

The cleanup reduces allocations. The samples do not establish a general latency improvement.
