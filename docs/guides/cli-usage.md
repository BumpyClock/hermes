# CLI usage

The Hermes CLI extracts content from one or more URLs.

## Installation

```bash
go install github.com/BumpyClock/hermes/cmd/hermes@latest
```

## Commands

- `hermes parse [flags] <url...>`: Parse one or more URLs
- `hermes version`: Print version information

## Flags

- `-f, --format <json|html|markdown|text>`: Output format. The default is `json`.
- `-o, --output <path>`: Write to file instead of stdout
- `--timeout <duration>`: Timeout per URL. The default is `30s`.
- `--concurrency <n>`: Maximum concurrent requests. The default is `10`. Use a positive integer.
- `--timing`: Print timing information to stderr

For one URL, `json` returns the full result with HTML content. Other formats return only the content in the selected format.

For multiple URLs, the CLI always returns a JSON array. Each entry contains `url`, `parseTime`, and `result`.
The format flag controls `result.content` in that array. The CLI omits failed URLs.

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

## Exit codes

- `0`: At least one URL parsed successfully and the CLI wrote the output.
- `>0`: The command failed. See stderr for errors.

## Tips

- Use `json` to retain the full `Result` structure for one URL.
- Use `--timing` to show elapsed times and errors for individual URLs.
- Use `markdown` or `text` to select the format of `Result.Content`.
