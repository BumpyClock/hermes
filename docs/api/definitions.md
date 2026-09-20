# Immutable local YAML definitions

```go
snapshot, err := hermes.LoadDefinitions("/opt/hermes-definitions/definitions")
if err != nil {
    return err
}
client := hermes.New(hermes.WithDefinitions(snapshot))
```

Loading is a separate, fallible, local-only operation. `New(opts ...Option) *Client`
and public `Result` JSON are unchanged. Share snapshots and clients across
goroutines. Loading copies all rules into private storage; construction and article
parsing never read definition files. Changes on disk require explicit loading and
new clients. A failed load returns no snapshot, even if other files are valid.

An explicitly configured client uses only its snapshot and generic fallback.
`WithDefinitions(nil)` and a zero-value `Definitions` select generic-only parsing.
Unconfigured clients are generic-only and perform no definition I/O.

## Local source and CLI

```sh
hermes parse --definitions /opt/hermes-definitions/definitions https://www.bbc.com/news/articles/example
```

The CLI loads once before starting article work. Missing, unreadable, empty or
invalid configuration fails with a nonzero exit and diagnostics on stderr, not
JSON/stdout. An explicitly empty `--definitions` is an error. The directory's
immediate `.yaml` and `.yml` files (case-insensitive extensions) are loaded; other
files and subdirectories without these extensions are ignored. A YAML-named
directory or symlink is rejected. There must be at least one rule file. Each
file contains exactly one complete site definition. Use a dedicated `definitions/`
directory, not the repository root.

## Schema 1

```yaml
schema: 1
site: example
hosts: [example.com, '*.news.example.com']
metadata:
  title:
    - text: h1
    - attribute: {selector: 'meta[name="og:title"]', name: value}
  author:
    - text: .byline
  date_published:
    - attribute: {selector: time, name: datetime}
  lead_image_url:
    - attribute: {selector: 'meta[name="og:image"]', name: value}
content:
  groups:
    - [article, '.article-body']
    - [main]
  remove: ['.recommendations', '.share']
  default_cleaner: true
```

Required fields are `schema`, `site`, `hosts` and `content.groups`. `metadata`,
its four fields, `content.remove`, `content.transforms`, and `content.default_cleaner` are optional.
All supplied sequences must be nonempty. Site identifiers are unique nonempty
strings. No inheritance, merging, aliases, anchors, duplicate keys, unknown
fields, coercion of strings to booleans/numbers, or additional YAML documents are
accepted. Unsupported operations and legacy options (including map-based transforms,
`allowMultiple`, date formats/timezones, and arbitrary expressions) are rejected
rather than partially executed. Bounded JSON sources and explicitly named
algorithms are documented in [ordered transforms](definition-transforms.md).

Metadata alternatives are explicitly either `{text: CSS}` or
`{attribute: {selector: CSS, name: ATTRIBUTE}}`. They read the first matching
element. Title, author and image alternatives stop at the first nonempty raw
value; date alternatives stop at the first successfully parsed date. Captured
dates accept ISO forms and unambiguous English month-first or day-first forms,
such as `September 17, 2026` and `17 September 2026`. Existing
field cleaners apply. Selectors run against Hermes's prepared document, where
meta `property` and `content` attributes become `name` and `value`. Attribute
names must match `[a-zA-Z_][a-zA-Z0-9_.:-]*`.

Content groups are ordered alternatives, not metadata tuples. Within a group,
selectors form a union in document order, with node deduplication. Selecting a
parent and its descendant includes the descendant only through the parent.
The first group with nonempty raw content wins, even when cleanup empties it;
missing extracted content can then use generic fallback, not a later group.
Ordered transforms run on selected content copies, followed by removal selectors
and then default cleaning. See [ordered transforms](definition-transforms.md)
for the typed operations, predicates, required-input policy and execution limits.
`default_cleaner` defaults to true. False skips heuristic/default cleanup but
does not bypass output sanitization (including Markdown/text conversion),
absolute URL resolution, URL validation, context cancellation or SSRF checks.
Relative links and selected lead-image metadata honor the source document's first base URL, resolved against the
article URL. URL normalization runs only on the selected output copy after
rule evaluation and removal, including when default cleaning is disabled.
Metadata, content and removal selectors therefore see the source's relative
attributes; generic fallback retains the same unmodified source.
Generic fallback supplies missing title/author/date/content;
remaining site metadata retains existing generic behavior.
Generic author fallback ignores site-wide sidebar or navigation profile links
that lack article/byline context. Explicit author metadata and article-local
author markup retain their existing priority.
Plain-text article output separates block elements, line breaks, list items and
table cells with whitespace while retaining inline text adjacency. This does not
change flat metadata tag stripping.

## Host matching

Hostnames are case-insensitive DNS names (maximum 253 bytes, labels up to 63).
Exact declared hosts with one leading `www.` equivalence win first. Otherwise
the longest matching explicit `*.example.com` suffix wins. Wildcards match any
subdomain depth on DNS-label boundaries, never the root. Exact/www ownership
collisions and duplicate wildcard ownership across sites reject the entire
load. Duplicate equivalent host claims within one site are harmless. Overlapping
different-specificity wildcards are allowed. There is no file-order precedence:
one complete definition is selected without merging.

## Diagnostics, limits and capability discovery

Use `errors.As(err, &definitionError)` with `*hermes.DefinitionError`.
`Source`, `Line`, `Column`, `Site`, and `Field` describe available context;
zero positions mean unavailable. `Unwrap` preserves the underlying YAML, CSS,
filesystem or validation error. Configuration errors are distinct from article
`ParseError` values.

`hermes.DefinitionCapabilities()` returns independent data describing schema 1,
implemented capabilities and these enforced limits:

| Limit | Value |
| --- | ---: |
| YAML files per directory | 1,024 |
| Bytes per file | 262,144 |
| Total input bytes | 16,777,216 |
| YAML nodes per file, including mapping keys | 10,000 |
| Node depth below document root (root = 0) | 16 |
| Items per sequence | 128 |
| UTF-8 bytes per scalar, including keys | 2,048 |

Comments count toward byte limits. Validation happens before activation.
Transform capabilities and their additional limits are documented in
[ordered transforms](definition-transforms.md). Capability IDs are stable strings;
consumers should compare membership, not depend on slice ordering.
These capabilities are the local runtime contract, not a promise of remote
bundle, cache or updater support. Canonical BBC and NYTimes YAML and synthetic provenance
live in `BumpyClock/hermes-definitions`; production site rules are not embedded.
