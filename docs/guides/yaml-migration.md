# Migrating to external YAML definitions

The external-only engine keeps the public Go module path, `New(opts ...Option)
*Client`, `Parser` interface, existing options, and `Result` JSON shape. It does
**not** preserve automatic selection of the compiled site rules shipped in
v1.1.1. Unconfigured library clients and CLI commands now use generic extraction
only, without reading definition files or downloading definitions. Their output
also differs from v1.1.1 as described under
[compatibility boundaries](#compatibility-boundaries).

Applications that depended on site rules must explicitly supply a snapshot.
The canonical corpus is maintained independently in
[`BumpyClock/hermes-definitions`](https://github.com/BumpyClock/hermes-definitions).
An accepted migration disposition can include a documented inaccessible-site
exception; it is not a claim that every current live site was tested successfully.

## Choose a startup source

1. For reproducible offline deployments, provision a reviewed, revision-pinned
   definitions directory alongside the application. Call `LoadDefinitions`
   before constructing clients, handle its error, then pass `WithDefinitions`.
2. For managed releases, call `LoadManagedDefinitions` with an exact `Version`
   or `Automatic: true`, not both. Handle fatal errors and surface `Warning`
   when a validated cache substitutes for a failed update check.
3. Share the resulting immutable snapshot across clients and workers. A file
   change or newly published release does not alter existing clients. Restart
   with a newly loaded snapshot to activate it.

```go
snapshot, err := hermes.LoadDefinitions("/opt/hermes-definitions/definitions")
if err != nil {
    return err
}
client := hermes.New(
    hermes.WithDefinitions(snapshot),
    hermes.WithContentType("markdown"),
)
```

For the CLI, add `--definitions DIRECTORY`, `--managed-definitions TAG`, or
`--managed-definitions-auto` to `hermes parse`. These sources are mutually
exclusive. `--definitions-cache DIRECTORY` is only for a managed source.
Missing, unreadable, empty, or invalid explicit sources fail before article
workers start; there is no silent substitution of generic-only configuration.

Managed mode requires an actual published compatible release. Local candidate
conformance does not publish one. Use a pinned local directory when no approved
release is available.

## Compatibility boundaries

Loaded definitions supply site selectors and bounded transforms; missing fields
retain generic fallback. Genuine transform failures return `ErrExtract` with a
contextual error chain rather than hiding the failure behind generic output.
Cancellation, timeout, and SSRF checks still apply. Turning off a definition's
default cleaner does not turn off sanitization.

Some output changes are deliberate repairs rather than byte-for-byte legacy
compatibility: Markdown preserves literal punctuation and entity-like text,
plain text separates block elements, generic authors exclude unrelated sidebar
profiles, and definitions preserve approved article structures and media.
Site acceptance evidence records selector repairs and inaccessible-site
exceptions separately. The ineffective Wikipedia transform and fabricated
YouTube placeholders are not a contract to reproduce.

Generic output also differs from v1.1.1 for every client, including clients
without a snapshot and CLI runs without a definition flag:

- When generic extraction finds no article content, `Content` holds page text
  from the [last-resort fallback](../api/hermes.md#parsing) instead of being
  empty. A missing title falls back to `<title>` or the first `<h1>`.
- Markdown and plain text are converted from sanitized HTML. Markdown links with
  a scheme other than `http` or `https`, such as `javascript:` or `mailto:`,
  render as their label or image, and non-ASCII URLs are percent-encoded as in
  HTML output.
- HTML output has no leading or trailing whitespace.

Do not treat a passing synthetic fixture as a fresh live-site observation or an
inaccessible-site exception as a successful scrape. Use the definitions
repository's cohort evidence and immutable candidate report for the exact
engine/corpus pair being deployed.

## Troubleshooting and rollback

| Outcome | Action |
| --- | --- |
| Generic output where a site rule was expected | Verify that the client received a snapshot and that its host matches a declared host or alias. There is no legacy registry fallback. |
| Definition loading fails | Inspect `DefinitionError` source, line, site, field, and underlying cause. Correct the source; do not continue as though it loaded. |
| Article transform fails | Inspect `ParseError` and `DefinitionOperationError`. Required inputs and resource limits fail explicitly. |
| Managed startup returns a warning | Record version, digest, update status, and warning. An intact cached snapshot was used; the update check did not succeed. |
| Managed startup fails without usable cache | Restore acquisition access or deliberately provision a validated local snapshot. No article work has started. |
| Rollback is needed | Restart with an older exact managed version or a previously approved local directory. Automatic mode is not a rollback pin. |

Checksums bind archive bytes to the trusted publisher's manifest; they do not
protect against a compromised publisher. Cache contents are revalidated at
startup. Neither loader merges sources, hot-reloads clients, or relaxes article
network policy.

See [local definitions](../api/definitions.md),
[ordered transforms](../api/definition-transforms.md),
[managed loading](../api/managed-definitions.md), and
[candidate conformance](../development/definition-conformance.md).
