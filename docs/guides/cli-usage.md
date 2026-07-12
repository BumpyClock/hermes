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
- `--timeout <duration>`: Timeout per URL (default: `30s`)
- `--concurrency <n>`: Maximum concurrent requests (default: `10`)
- `--timing`: Print timing information to stderr

Notes:

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

## Exit Codes

- `0`: At least one URL parsed successfully
- `>0`: All URLs failed to parse (see stderr for errors)

## Tips

- Prefer `json` for batch operations; it includes the full `Result` structure.
- Use `--timing` to understand performance characteristics across multiple URLs.
- When using `markdown` or `text`, the parser returns `Result.Content` in that format.
