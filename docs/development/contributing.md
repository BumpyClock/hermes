# Contributor guide

Use this guide to prepare code, tests, and documentation for a pull request.

## Prerequisites

- The Go version in [`go.mod`](../../go.mod).
- Git.
- Make for the commands below.
- `golangci-lint` v2.12.2, built with support for the selected Go toolchain.

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

## Prepare a change

1. For a substantial change, open an issue to discuss the problem and scope.
2. Create a branch with a descriptive name, such as `feat/custom-extractor` or `fix/date-parser`.
3. Change the code and add tests for the affected behavior.
4. Update the affected documentation and examples.
5. Submit a pull request with the validation results.

### Commit messages

Use `type(scope): subject` for commit subjects. Use an imperative subject without a final period.

The types are `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, and `chore`.

```text
feat(parser): add support for custom timeout configuration
fix(extractors): handle malformed JSON in custom extractor definitions
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
| `make verify` | Run lint, race tests with coverage, and the CLI build. |
| `make benchmark` | Run benchmarks with allocation metrics, without ordinary tests. |
| `make build` | Build the CLI at `bin/hermes`. |

Test observable behavior and regression risks, not each function or implementation detail.
For performance changes, compare benchmarks with the same inputs and commands.

```bash
make benchmark PACKAGES=./internal/parser BENCH=BenchmarkParseHTML BENCHFLAGS='-count=5'
```

Before a pull request, run `make verify`. Report failures and omitted checks.
All required checks must pass before merge. Do not reduce test coverage without an explanation.

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
