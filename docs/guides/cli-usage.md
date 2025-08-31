# CLI Usage

Hermes includes a CLI for parsing one or more URLs and emitting results in multiple formats.

## Installation

```bash
go install github.com/BumpyClock/hermes/cmd/hermes@latest
```

## Commands

- `hermes parse [flags] <url...>`: Parse one or more URLs
- `hermes version`: Print version information

## Flags

- `-f, --format <json|html|markdown|text>`: Output format (default: `json`)
- `-o, --output <path>`: Write to file instead of stdout
- `--headers <json>`: Custom HTTP headers as JSON string (currently ignored)
- `--timeout <duration>`: Timeout per URL (default: `30s`)
- `--concurrency <n>`: Maximum concurrent requests (default: `10`)
- `--timing`: Print timing information to stderr

Notes:

- The `--headers` flag is parsed but not applied yet; only the `User-Agent` can be set via the library using `WithUserAgent`.
- The output format controls both CLI output and the content format requested from the parser.

## Examples

### Single URL

```bash
# JSON (default)
hermes parse https://example.com/article

# Markdown
hermes parse -f markdown https://example.com/article

# HTML
hermes parse -f html https://example.com/article

# Plain text
hermes parse -f text https://example.com/article
```

### Multiple URLs

```bash
# Parse two URLs and emit a JSON array
hermes parse https://example.com/1 https://example.com/2

# With timing output
hermes parse --timing https://example.com/1 https://example.com/2

# Save to file
hermes parse -o output.json https://example.com/1 https://example.com/2
```

### Custom Headers (currently ignored)

```bash
hermes parse --headers '{"User-Agent":"Hermes/1.0"}' https://example.com/article
# Note: headers are currently ignored by the CLI; use the Go API WithUserAgent option.
```

## Exit Codes

- `0`: At least one URL parsed successfully
- `>0`: All URLs failed to parse (see stderr for errors)

## Tips

- Prefer `json` for batch operations; it includes the full `Result` structure.
- Use `--timing` to understand performance characteristics across multiple URLs.
- When using `markdown` or `text`, the parser returns `Result.Content` in that format.
