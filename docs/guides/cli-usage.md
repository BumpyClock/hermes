# CLI Usage

Hermes includes a CLI for parsing one or more URLs and emitting results in multiple formats.

## Installation

```bash
go install github.com/BumpyClock/hermes/cmd/parser@latest
```

## Commands

- `parser parse [flags] <url...>`: Parse one or more URLs
- `parser version`: Print version information

## Flags

- `-f, --format <json|html|markdown|text>`: Output format (default: `json`)
- `-o, --output <path>`: Write to file instead of stdout
- `--headers <json>`: Custom HTTP headers as JSON string
- `--fetch-all`: Attempt to handle multipage articles (default: true)
- `--timing`: Print timing information to stderr

Note on pagination: the parser detects `next_page_url` when present. Automatic fetching and merging of paginated articles is not yet implemented; use the `next_page_url` field to drive your own pagination logic if needed.

## Examples

### Single URL

```bash
# JSON (default)
parser parse https://example.com/article

# Markdown
parser parse -f markdown https://example.com/article

# HTML
parser parse -f html https://example.com/article

# Plain text
parser parse -f text https://example.com/article
```

### Multiple URLs

```bash
# Parse two URLs and emit a JSON array
parser parse https://example.com/1 https://example.com/2

# With timing output
parser parse --timing https://example.com/1 https://example.com/2

# Save to file
parser parse -o output.json https://example.com/1 https://example.com/2
```

### Custom Headers

```bash
parser parse --headers '{"User-Agent":"Hermes/1.0","Accept":"text/html"}' \
  https://example.com/article
```

## Exit Codes

- `0`: At least one URL parsed successfully
- `>0`: All URLs failed to parse (see stderr for errors)

## Tips

- Prefer `json` for batch operations; it includes the full `Result` structure.
- Use `--timing` to understand performance characteristics across multiple URLs.
- When using `markdown` or `text`, the CLI converts the `content` field accordingly before writing.

