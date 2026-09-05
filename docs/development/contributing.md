# Contributor guide

Use this guide to prepare code, tests, and documentation for a pull request.

## Prerequisites

- The Go version in [`go.mod`](../../go.mod).
- Git.
- Make for the commands below.
- `golangci-lint` for lint checks.
- `gofumpt` for code format checks.

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

Check code format:

```bash
gofumpt -d .
```

Apply format changes only to files in the scope of your change:

```bash
gofumpt -w path/to/changed.go
```

Run the linter:

```bash
make lint
```

`make lint` runs `golangci-lint run`.

## Tests

Use table-driven tests for related cases. Test individual functions and component interactions with fixtures or local HTTP servers where possible.

Add regression tests for bug fixes. Add benchmarks with allocation metrics for performance-sensitive changes.

Keep live-site tests separate from deterministic tests. A remote site can change content or reject requests without a code change.

Choose checks that cover the affected behavior:

| Command | Purpose |
| --- | --- |
| `make test` | Run all tests with coverage. |
| `go test ./... -short` | Enable short mode. Tests must check `testing.Short()` to skip work. |
| `go test -race -coverprofile=coverage.out ./...` | Run the test command from CI. |
| `make benchmark` | Run tests and benchmarks with allocation metrics. |
| `make build` | Build the CLI at `bin/hermes`. |

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

1. Update your branch from `upstream/main`.
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
