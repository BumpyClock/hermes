# Contributor guide

Use this guide to prepare code, tests, and documentation for a pull request.

## Prerequisites

- The Go version in [`go.mod`](../../go.mod).
- Git.
- Make for the commands below.
- `golangci-lint` v2.13.2, built with support for the selected Go toolchain.
- Python 3.11+ and Git for offline tooling and hook fixtures. CI uses Python 3.12.3.

For deterministic Go/linter preparation, install Go 1.27.1 on `PATH`, then run:

```bash
GOTOOLCHAIN=go1.27.1 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2
go mod download
```

The explicit `GOTOOLCHAIN` matters. An older bootstrap compiler can otherwise build the linter with a Go version too old for this module.

## Prepare a local checkout

1. Fork the repository on GitHub.
2. Clone your fork:

   ```bash
   git clone https://github.com/YOUR_USERNAME/hermes.git
   cd hermes
   ```

3. Add the upstream remote:

   ```bash
   git remote add upstream https://github.com/BumpyClock/hermes.git
   ```

4. Install dependencies:

   ```bash
   make deps
   ```

5. Check the checkout:

   ```bash
   make test
   make lint
   ```

`make deps` runs `go mod download` and `go mod tidy`. It can change `go.mod` and `go.sum`.

## Opt-in pre-push validation

After preparing dependencies and committing the tracked hook files, explicitly opt in:

```bash
./scripts/install-hooks.sh
make test-hooks
```

The installer is idempotent, works in linked worktrees, and refuses unrelated active hooks or `core.hooksPath`. It changes only repository-local configuration. It is never run automatically. Default hooks behind symlinked directories are preserved. Narrow `.gitattributes` rules keep only the shell entrypoint and installer on LF under `core.autocrlf`; Git-status errors fail closed.

The hook requires a clean checkout, including untracked files, and validates only refs resolving to its checked-out `HEAD`. Other pushed refs fail with an actionable error; use their own clean checkout. Multiple refs, new branches, missing remote objects, renames and deletions are handled without fetching or stashing. Empty input or an empty diff broadens validation; deletion-only pushes do not run project checks.

Changed existing Go files use configured `golangci-lint fmt --diff`, with any diff treated as failure. Shared configuration, dependency changes or uncertain paths broaden formatting to all tracked Go files. The hook always runs configured full-module lint, `go vet ./...`, `go build ./...` and offline release/compatibility tests. A reverse-dependency fast path is deliberately not claimed. All nonzero exits propagate. No source rewriting, automatic dependency/toolchain downloads, external definition downloads, live website checks, release operations or credential use occur. Prepare missing tools/dependencies explicitly; module access is offline and read-only during the hook.

CI retains full Go race/coverage and CLI/benchmark checks. Fixture tests use stub tools and disposable Git repositories, not outbound pushes. Their pass does not prove compiler or linter acceptance. Hook latency depends on the prepared tool cache and machine; no nine-second budget is promised.

## Prepare a change

1. For a substantial change, open an issue to discuss the problem and scope.
2. Create a branch with a descriptive name, such as `feat/custom-extractor` or `fix/date-parser`.
3. Change the code and add tests for the affected behavior.
4. Update the affected documentation and examples.
5. Submit a pull request with the validation results.

### External definition changes

The engine is generic-only by default. Add or revise site behavior in a
separately loaded external YAML snapshot and its deterministic fixture evidence,
not in compiled Go site factories. Shared bounded operations or named
exceptions require engine tests and capability documentation; snapshot loading
remains a fallible startup step.

### Commit messages

Use `type(scope): subject` for commit subjects. Use an imperative subject without a final period.

The types are `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, and `chore`.

```text
feat(parser): add support for custom timeout configuration
fix(definitions): handle malformed JSON in external definition rules
docs(api): update parser configuration examples
```

Use the commit body to explain reasons that the subject and diff do not show.

## Code standards

Keep non-public code in focused packages under `internal/`. Match the names, import groups, and structure of the affected package.

Document exported functions and types with complete comments. Explain constraints and reasons that the code does not express.

Include context in errors. Preserve the underlying error with `%w` when callers need to inspect it.
Use the public `ParseError` fields `Code`, `URL`, `Op`, and `Err` for structured errors.

Use descriptive names. Avoid abbreviations that hide the role of a value.

### Format and lint checks

Check code format with the formatters in [`.golangci.yml`](../../.golangci.yml):

```bash
golangci-lint fmt --diff
```

Apply format changes only to files in the scope of your change:

```bash
golangci-lint fmt path/to/changed.go
```

Run the linter:

```bash
make lint
```

`make lint` runs `golangci-lint run`.

## Tests

Use table-driven tests for related cases. Test individual functions and component interactions with fixtures or local HTTP servers where possible.

Add regression tests for bug fixes. Add benchmarks with allocation metrics for performance-sensitive changes.

Use local fixtures, injected transports, or `httptest` servers instead of public network requests.
The suite has no separate short-mode or integration build-tag tier.
See [`context_test.go`](../../context_test.go) and [`integration_test.go`](../../integration_test.go) for executable examples.

Choose checks that cover the affected behavior:

| Command | Purpose |
| --- | --- |
| `make test` | Run all tests with coverage. |
| `go test ./internal/parser -run TestName -count=1` | Run a focused regression test. |
| `make test-race` | Run the race and coverage checks from CI. |
| `make test-release` | Run offline release-tool tests with Python 3.11 or later. |
| `make test-compatibility` | Test the released-version comparison and exact-difference acceptance rules. |
| `make test-hooks` | Test executable hook and installer behavior with stub tools in disposable Git repositories. |
| `make verify` | Run lint, race tests with coverage, the CLI build, release/compatibility-tool tests and hook fixtures. |
| `make benchmark` | Run benchmarks with allocation metrics, without ordinary tests. |
| `make build` | Build the CLI at `bin/hermes`. |

Test observable behavior and regression risks, not each function or implementation detail.
For performance changes, compare benchmarks with the same inputs and commands.

```bash
make benchmark PACKAGES=./internal/parser BENCH=BenchmarkParseHTML BENCHFLAGS='-count=5'
```

Before a pull request, run `make verify`. Report failures and omitted checks.
All required checks must pass before merge. Do not reduce test coverage without an explanation.

Compatible Linux CI, benchmark-script, and Claude jobs use Boat. The race step drops inherited/ambient `CAP_DAC_READ_SEARCH` with `setpriv` before `make test-race`. This preserves unreadable-file assertions under ordinary user permissions instead of skipping them; it changes only the test process, not account settings or test coverage. Runner-label changes do not fix provider API billing or authorize enabling public execution. Public execution follows the separately configured actor allowlist; outsiders and bots are not implicitly authorized. A successful live run remains an external acceptance gate. Do not change account settings or credentials as part of local validation.

CI, benchmark checks, and automatic PR reviews cancel older runs for the same workflow and branch or pull request.
Separate workflows and branches do not cancel each other.
Comment-triggered Claude requests remain independent because separate comments can contain distinct tasks.
All workflow jobs have explicit time limits.
The benchmark workflow retains its path filters.

See [the release procedure](../RELEASING.md) for release gates and recovery after partial publication.

## Documentation

Keep `doc.go` package comments and exported API comments consistent with the public API.
Provide examples that compile when you add or change public behavior.

Use the relevant section of `docs/`:

- [API reference](../api/hermes.md) for types, methods, and options.
- [User guides](../guides/basic-usage.md) for tasks and procedures.
- [Examples](../examples/basic.md) for sample code.
- [Architecture](../architecture/overview.md) for internal behavior and package roles.

Check links, commands, and identifiers against the current source.

## Pull requests

### Before submission

1. Update your branch from `upstream/master`.
2. Review the diff for unrelated changes.
3. Run the tests, lint checks, and build checks that apply to the change.
4. For performance changes, record benchmark results.
5. State compatibility changes and unresolved failures in the pull request.

### Pull request description

Use a short title with the same `type(scope): subject` format as commit subjects.
Explain the problem before the implementation. Include these sections when relevant:

```markdown
## Why
Describe the problem and why the change is necessary.

## Scope
Name the affected behavior, packages, and compatibility changes.

## Tradeoffs
Explain material decisions and limitations.

## Verification
List exact commands and results. State omitted checks and their reasons.

## Related issues
Closes #<issue_number>
```

Maintainers review the code and documentation before approval. CI checks tests, lint, the CLI build, and benchmarks.

## Issues and questions

Search the documentation and existing issues before you open an issue.

For a bug report, include:

- The expected and actual behavior.
- Exact reproduction steps or a minimal Go example.
- The operating system, `go version` output, and Hermes version.
- Relevant errors and a URL or HTML fixture, with secrets and personal data removed.

For a feature request, describe the problem, proposed behavior, alternatives, and a concrete use case.

Keep discussion respectful. Ask for clarification when a requirement is unclear, and provide evidence for technical claims.

The project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).
