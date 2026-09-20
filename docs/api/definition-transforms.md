# Ordered YAML transforms

Schema 1 accepts a nonempty `content.transforms` sequence. Each step has an
explicit `target`, exactly one operation mapping, and optionally `when`.
Unsupported operations or parameters fail `LoadDefinitions`; they never become
silently ignored settings. Operations use parsed DOM nodes and typed scalar
values only. They cannot execute scripts or expressions, interpolate HTML,
perform general JSON queries, read files, make requests, or access a shell.

```yaml
content:
  groups: [[article]]
  transforms:
    - target: descendants
      selector: img.g-lazy
      string.replace: {attribute: src, old: '{{size}}', new: '640'}
    - target: descendants
      selector: 'img[src$="-640.jpg"]'
      when:
        - exists: {attribute: src}
        - number: {attribute: width, comparison: ge, value: 400}
      attribute.set: {name: alt, value: Article photograph}
    - target: descendants
      selector: .caption
      element.rename: {tag: figcaption}
  remove: [.advertisement]
  preserve: [figure.article-image]
  default_cleaner: true
```

## Ordering and ownership

1. Select the first nonempty content group from the prepared source, retaining
   source-order deduplication and outermost selected roots.
2. Check the selected content's resource limits and copy it into an isolated
   container. Execute all steps against this combined copy in YAML order.
3. Snapshot each step's matching elements once, in document order. Process deeper
   elements first, with stable document-order ties. Check conditions immediately
   before each operation, so a parent's predicate observes earlier child changes.
4. Query the mutated copy afresh for the next step. A node may receive several
   deliberate steps. Never repeatedly select newly matching nodes within a step.
5. Apply site removals, output URL normalization, default cleaning when enabled,
   and existing output sanitization/format conversion.

`content.preserve` is a nonempty sequence of CSS selectors evaluated on the
transformed content. It protects matching elements and their descendants from
the default cleaner's image, empty-node, heading, and conditional-content
heuristics. It does not affect explicit `remove` selectors, raw/executable tag
stripping, URL normalization, attribute cleaning, or final sanitization.

## Metadata text capture

Metadata selector alternatives also support a bounded `text_capture` mapping:

```yaml
metadata:
  date_published:
    - text_capture:
        selector: .author
        pattern: '([0-9]{4}-[0-9]{2}-[0-9]{2})'
        group: 1
```

It reads the selected source DOM without mutating it, bounds collected text to
`MaxValueBytes` and traversal to the content node/depth limits, then returns
the named RE2 capture. `group` starts at one and must exist at load time. A
selector or regex miss is an ordinary empty alternative, so later metadata
alternatives still run; an invalid captured date also continues to the next
date alternative. Invalid patterns/groups reject the definition, and runtime
resource or cancellation failures return contextual extraction errors.

Date values use the engine's supported date layouts, including year-first
slash-separated dates with padded or unpadded months and days (`2019/03/05`
or `2019/3/5`) and datetimes with an explicit UTC suffix
(`2026-09-19 02:21:00 UTC`). Year-first Japanese dates such as `2019年4月4日`
also retain their stated year, month, and day. This does not add per-definition format or timezone
settings, or interpret arbitrary timezone abbreviations as UTC.

`target: root` selects the original copied content roots and forbids `selector`.
`target: descendants` requires a valid CSS `selector` and excludes those roots.
Unwrapping retires a root; its promoted children remain descendants. Removed roots
do not become targets again. Renaming keeps node identity, attributes and children.
Unwrapping moves children in sibling order. Detached matches are skipped safely.
The synthetic container is never a rule target and is not serialized.
With transforms, surviving roots are passed to output cleaning too, so root
renames are observable when the resulting tag is sanitizer-approved.

Metadata and generic fallback still read the prepared source, not the mutated
copy. Relative attribute selectors see source values until an explicit
`url.resolve` or other earlier step changes them. Output resolution still honors
the first source `base[href]`; transformations cannot replace that base context.
Disabling default cleaning does not disable sanitization.

This intentionally differs from the retired Go executor: YAML does not
alphabetically sort selector keys or globally deduplicate nodes across steps.
Simple renames use the same typed executor as every other YAML operation, not
the legacy string/HTML replacement implementation. Unconfigured clients use
generic extraction only; there is no compiled-site fallback.

## Operations

Every parameter not marked optional is required. Attribute names are validated
and normalized to lowercase; values are literal strings, with no interpolation.
The existing attribute-name grammar also applies here.

| Operation key | Parameters and behavior |
| --- | --- |
| `attribute.copy` | `from`, `to`, optional `required`. Copy one attribute on the matched element. |
| `attribute.set` | `name`, `value`. Set a literal attribute, including an empty value. |
| `attribute.set_from` | `name`, `value` source. Read a bounded structured value and set it on the matched element; an absent optional source leaves the attribute unchanged. |
| `attribute.remove` | `name`. Remove an attribute; absence is normal. |
| `string.replace` | `attribute`, `old`, `new`, optional `required`. Replace every literal occurrence. `old` must have at least one byte (spaces are valid); `new` may be empty. |
| `regex.capture` | `attribute`, `pattern`, integer `group`, optional `required`. Replace the attribute with the first match's specified capture. Group numbers start at 1 and must exist. Go's RE2-syntax regular expressions reject backreferences and lookaround. |
| `url.resolve` | `attribute`, optional output `to`, optional typed `base`, and optional `required`. Resolve the source attribute against the explicit base source or the first source base/article URL; require an HTTP(S) result without credentials. |
| `url.build` | `attribute`, `base`, optional `path` and `query`. Construct a bounded HTTP(S) URL as described below, then set the attribute. |
| `element.rename` | `tag`. Rename in place using DOM fields; never parse generated HTML. |
| `element.unwrap` | Empty mapping `{}`. Remove only the element, retaining its children. |
| `element.remove` | Empty mapping `{}`. Remove the element and its subtree. |
| `element.retain` | `selector`, optional `select: first\|last\|all` (default `all`), optional `required`. Keep only the selected descendant elements, in source order, below each matched step target. Nested duplicate selections retain their outermost element once. |
| `element.move` | `source` and `target` mappings, each with `selector` and optional `select: first\|last\|all`; `position: append\|prepend\|before\|after\|replace`; optional `required`. Move selected descendant sources to explicit descendant targets. Multiple targets receive source-order clones after the first target so `all` has deterministic, non-lossy behavior. |
| `element.create` | `target` mapping with `selector` and optional `select`, or exactly `{self: true}` for the current step node; `position`; and `node` with a safe `tag`, optional literal `text`, and literal string `attributes`. The node is constructed through DOM APIs for every selected target. |
| `noscript.recover` | `source` and `target` selection mappings, `position`, and optional `required`. Recover `img` elements from parsed-child or entity-encoded `noscript` content, insert DOM clones at the explicit targets, then remove the source `noscript` nodes. |
| `algorithm.apply` | `name: abendblatt.deobfuscate`. Run the explicitly named, built-in Abendblatt character decoder on an `obfuscated` matched element. Unknown names are load errors. |

Selection mappings snapshot descendants in document order. `first`, `last`, and
`all` apply to that snapshot; a later YAML step sees mutations from earlier
steps. Missing optional move, retain, or recovery sources/targets are normal
no-ops. `required: true` turns their absence, or recovery with no image, into a
contextual extraction error. A move cannot insert a node into itself or one of
its descendants; such cycles are extraction errors. Detached targets are skipped
by the step snapshot and detached insertion targets are errors rather than
causing a panic.

`element.create` alone also accepts `target: {self: true}`. It applies the
requested insertion or replacement to the current step node without requiring
a fabricated child selector; no `select` field is permitted. Replacing a
current content root retires that root under the existing root semantics, while
later steps continue to see its surviving replacement through normal descendant
queries.

`noscript.recover` also accepts `{self: true}` for both `source` and `target`,
with `position: replace` only. Applied to a selected `noscript` step node, it
replaces that node with only its own recovered images in source order. This
avoids cartesian cloning when a content root has several noscript elements.

Rename tags are explicitly limited to:
`a article aside b blockquote caption code dd del div dl dt em figcaption figure
h1 h2 h3 h4 h5 h6 i li ol p pre s section span strong table tbody td tfoot th
thead tr u ul`. Constructed tags add only `br` and `img`; `img` cannot contain
text. Void, executable, raw-text and noscript target tags are not accepted.
Sanitization may still discard tags or attributes in the output.

### Structured value sources

`attribute.set_from` values and `element.create.node.attributes` accept one
typed source. A string value is a literal. The existing `{attribute: NAME,
required: BOOL}` source reads the matched element. A
`{descendant_attribute: {selector: CSS, name: ATTRIBUTE, select:
first|last, required: BOOL}}` source reads one explicit descendant attribute.

A JSON source has no filters, expressions, recursive descent, or dynamic keys:

```yaml
value:
  json:
    attribute: data-props
    path: [{field: sources}, {index: 0}, {field: src}]
    required: true
```

It parses the selected attribute as one JSON value, walks only the listed field
and nonnegative-index segments, and returns a final JSON string. Missing
optional inputs, malformed optional JSON, missing fields, wrong optional types,
and optional out-of-bounds indexes skip the consuming operation. Required
variants return a contextual extraction error. Attribute values and constructed
elements are assigned through DOM APIs; extracted JSON is never interpolated as
HTML.

The only named algorithm is `abendblatt.deobfuscate`. It is selected explicitly
by YAML rather than hostname, site configuration, or a general algorithm
registry. It replaces decoded content with a DOM text node, so hostile decoded
characters remain text and final sanitization still applies.

`required` defaults to false for attribute inputs. Missing optional inputs skip
the operation or fail the predicate normally. Missing required attributes,
empty/whitespace-only required values, required captures that do not match or
capture only whitespace, malformed present numeric values, invalid URLs and
resource-limit violations fail extraction. An optional regex with no matching
capture leaves the attribute unchanged. An unmatched selector always skips the
step, even when the operation would require an input.

### URL construction

`base` must be a load-time-valid absolute HTTP(S) URL without credentials and
with valid query encoding. Each path item and query value is exactly one typed
source: `{literal: STRING}` or `{attribute: NAME, required: BOOL}` (the boolean
is optional). Missing optional sources skip the entire build without partially
modifying the target attribute.

```yaml
- target: descendants
  selector: a.media
  url.build:
    attribute: href
    base: https://media.example.com/items?format=image
    path:
      - literal: photo
      - attribute: data-id
        required: true
    query:
      - name: size
        value: {literal: '640'}
      - name: caption
        value: {attribute: data-caption}
```

Path items append individually escaped segments to the base path. Empty, `.` and
`..` segments are rejected; slashes in literal or extracted values are escaped as
`%2F`, not treated as separators. Escaped slashes in earlier segments or the base
remain escaped when later segments are appended: `['a/', 'b']` produces
`/a%2F/b`, not `/a/b`. Query names must be unique within the declaration;
values replace same-named base parameters and are URL-encoded. Base query
parameters not explicitly replaced remain. No operation fetches any URL, reads
files or changes network policy. Article URL validation and SSRF protections
remain independent.

`url.resolve.base` accepts the same literal or attribute source form as URL
construction. A supplied base is first resolved against the source document's
base/article URL, then must be an absolute credential-free HTTP(S) URL. Use
`to` when the reference and destination attributes differ:

```yaml
url.resolve:
  attribute: data-original
  to: src
  base: {attribute: src, required: true}
```

## Conditions

`when` is a nonempty sequence of up to 16 conditions, evaluated in sequence
with AND semantics and short-circuiting. Each item has exactly one condition
mapping. There are no nested logical expressions.

| Condition | Parameters |
| --- | --- |
| `exists` | `attribute`; optional `present` (default true). Empty attributes still exist. |
| `text` | `comparison: equals\|contains\|prefix\|suffix`, `value`. Optional `attribute` reads that attribute instead of concatenated descendant DOM text; `required` is allowed only with `attribute`. Text is literal, case-sensitive and not whitespace-normalized. |
| `number` | `attribute`, `comparison: eq\|ne\|lt\|le\|gt\|ge`, finite YAML number `value`; optional `required`. Trim source whitespace and parse a finite float64. Invalid present values are errors, not unmet predicates. |
| `descendant` | CSS `selector`; optional `present` (default true). Excludes the current element. |

Use a false `present` for a supported absence test. String booleans and numeric
strings in declarations are not coerced.

## Errors and bounded work

Genuine operation failures propagate through both `Parse` and `ParseHTML` as
`*hermes.ParseError` with `Code: ErrExtract`, the original URL and original
operation. They do not silently trigger generic output.
`errors.As(err, &operationError)` with `*hermes.DefinitionOperationError` exposes
`Source`, `Site`, one-based `Step`, `Operation`, `Line`, `Column` and `Err`.
`Unwrap` preserves the cause, including numeric and URL parser errors.
`errors.Is(err, hermes.ErrDefinitionTransformLimit)` identifies resource failures.
Cancellation/deadline causes remain inspectable and retain the existing timeout
classification instead of becoming extraction failures.

| `DefinitionCapabilities()` field | Enforced maximum |
| --- | ---: |
| `MaxTransformSteps` | 128 steps per content program |
| `MaxConditions` | 16 conditions per step |
| `MaxPatternBytes` | 512 regex UTF-8 bytes |
| `MaxCaptureGroups` | 16 regex captures |
| `MaxValueBytes` | 65,536 bytes per consumed/produced scalar or text predicate |
| `MaxContentNodes` | 32,768 selected content nodes, including text and synthetic container |
| `MaxContentDepth` | 256 edges below the synthetic container |
| `MaxTransformWork` | 2,097,152 cumulative work units per selected content program |
| `MaxJSONDepth` | 16 JSON container levels |
| `MaxJSONTraversal` | 32 field/index segments |

One work unit is charged per visited/matched node, inspected initial data byte,
attribute key/value byte (plus one per attribute), or consumed/produced scalar
byte. Attribute writes also charge their produced key bytes and per-attribute
overhead before mutation, including empty values. Attribute reads, writes, and
removals charge every inspected attribute even when an optional input is
absent. Repeated steps and condition traversal consume the same budget. Content
size/depth checks precede cloning. Replacement expansion is checked before
allocation. Regex inputs are capped and use Go's linear-time regex engine.
Cancellation is checked during tree visits, between matched nodes/conditions,
and on scalar reads/writes. Shape-changing steps recheck generated node count
and depth before the next step. Scalar-only operations cannot change tree shape;
their bounded input/output reads and writes are charged directly, avoiding
redundant full-tree byte rescans for repeated renames or attribute edits.
Noscript text is bounded before HTML fragment parsing and only image DOM nodes
are recovered. CSS matching and regex calls themselves are not preemptible;
their inputs are bounded by these limits and the loader's existing 2,048-byte
scalar/CSS limit.

The node and work ceilings are calibrated for complete normal article roots,
including headings, lists, figures and tables, rather than requiring lossy
paragraph-only selection to fit an earlier 10,000-node cap. They remain finite:
over-limit trees and work still fail explicitly, and no site receives a special
exception. Synthetic boundary cases cover both the normal full-article range
and hard failures above these ceilings.

`DefinitionCapabilities().Capabilities` contains operation and condition IDs:
`transform.` plus each operation key in the table, and
`condition.exists`, `condition.text`, `condition.number`, `condition.descendant`.
Optional schema extensions have separate capability IDs:
`transform.url.resolve.base`, `transform.url.resolve.to`,
`element.create.typed_attributes`, `element.create.self_target`,
`transform.noscript.recover.self`,
`value.descendant_attribute`, and
`value.json`. These distinguish engines that implement the original operation
names but not their later typed fields.
`DefinitionCapabilities().Algorithms` independently contains the supported named
algorithm IDs, currently `abendblatt.deobfuscate`;
the cleaner control is `content.preserve`.
`DefinitionCapabilities()` returns new slices each call. Unsupported capability
requirements should be rejected by bundle consumers before loading; the local
loader itself rejects unsupported operation/condition keys and parameters.

Canonical NYTimes fixtures and provenance live in `hermes-definitions`.
Engine regressions use original synthetic inputs and public IP URLs, with an
injected transport for `Parse`. No live-site requests are needed.

The issue #34 validation passed the focused definition/transform/legacy-content
suite and `make verify` (lint, full race/coverage suite, CLI build and 17 offline
release-tool tests). The opt-in `TestCanonicalDefinitionsNYTimesOffline` gate also
passed with race detection for both public entry points and all three formats.
Its host-only IP mapping is explicit; it does not claim live NYTimes verification.
