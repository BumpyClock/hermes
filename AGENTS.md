# Hermes agent guide

Hermes extracts structured article data from HTML and URLs. The public Go module is `github.com/BumpyClock/hermes`.

## Source ownership

- `client.go`, `options.go`, `result.go`, and `errors.go` define the public API. The public result type is `Result`.
- Root `parser.go` defines the public `Parser` interface. `internal/parser/` coordinates extraction.
- `internal/extractors/custom/` contains site-specific extractors and the registry in `index.go`.
- `internal/extractors/generic/`, `internal/cleaners/`, and `internal/utils/` contain fallback extraction and content transforms.
- `cmd/hermes/` owns the CLI. `examples/` contains API consumers.
- `internal/fixtures/` contains local HTML inputs. `benchmark/` contains the separate JavaScript comparison tool.

## Verification

Use the Go version from `go.mod` and golangci-lint v2.13.2 with `.golangci.yml`.
The linter build must support the selected Go toolchain.

| Scope | Command |
| --- | --- |
| Focused regression | `go test ./path/to/package -run TestName -count=1` |
| Package coverage | `make test PACKAGES=./internal/parser` |
| Full local acceptance | `make verify` |
| Race checks with coverage | `make test-race` |
| CLI build | `make build` |
| Lint | `make lint` |
| Format changed Go files | `golangci-lint fmt path/to/file.go` |
| Repeated parser benchmark | `make benchmark PACKAGES=./internal/parser BENCH=BenchmarkParseHTML BENCHFLAGS='-count=5'` |

`make verify` runs lint, race tests with coverage, and the CLI build.
`make benchmark` excludes ordinary tests.
The suite has no separate `-short` or `integration` build-tag tier.
CI covers pushes to `master` and `develop`, plus pull requests that target `master`.

Test observable contracts and regression risks.
Use local fixtures, injected HTTP transports, or `httptest` servers for deterministic tests.
Assert the expected error code for denial tests.
For performance claims, compare repeated measurements with identical fixtures, commands, and environments.
Report exact checks, failures, and omitted checks.

## Contracts and documentation

Read [API documentation](docs/api/hermes.md) for public API changes.
Read [extractor documentation](docs/api/extractors.md) for custom extractor changes.
Read [contributor guidance](docs/development/contributing.md) for test and contribution conventions.
Update the affected documentation when behavior or public API contracts change.

Register new custom extractors in `GetAllCustomExtractors`.
Test site selectors and transforms against local fixtures.
Preserve SSRF denial behavior and context propagation across HTTP requests.
The public `Result` does not expose the internal pagination URL.

## Releases

Use [.agents/skills/create-release/SKILL.md](.agents/skills/create-release/SKILL.md) for an authorized release.
The release branch is `master`.
