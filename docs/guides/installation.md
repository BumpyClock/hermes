# Installation & Setup Guide

This guide covers installing the Hermes CLI and using Hermes as a Go library.

## Table of Contents

- System Requirements
- Installation Methods
- Quick Start
- Environment Setup
- Verification
- Troubleshooting

## System Requirements

- Go 1.24.6 or later (matching `go.mod`)
- Linux, macOS, or Windows

## Installation Methods

### Go Install (Recommended)

```bash
go install github.com/BumpyClock/hermes/cmd/hermes@latest
```

Installs the `hermes` CLI into `$GOPATH/bin`.

### Build from Source

```bash
git clone https://github.com/BumpyClock/hermes.git
cd hermes
make build
make install
```

Note: Pre-built binaries and package managers are not currently published.

## Quick Start

Verify CLI:

```bash
hermes version
```

Parse a URL:

```bash
hermes parse https://example.com/article
```

Use as a library:

```go
package main

import (
    "context"
    "fmt"
    "github.com/BumpyClock/hermes"
)

func main() {
    client := hermes.New()
    res, _ := client.Parse(context.Background(), "https://example.com")
    fmt.Println(res.Title)
}
```

## Environment Setup

Ensure Go is set up correctly:

```bash
go version
go env GOPATH
```

Hermes does not use environment variables for configuration. Configure via code options (see [Configuration](../api/configuration.md)).

## Verification

```bash
hermes parse -f markdown https://www.nytimes.com/section/technology
hermes parse --timing https://example.com
hermes parse --concurrency 5 https://example.com/1 https://example.com/2
```

## Troubleshooting

- Increase timeout: use `hermes.WithTimeout(60 * time.Second)` in the Go API.
- Reduce concurrency: use `--concurrency` for the CLI when parsing many URLs.
- SSL issues: update system CA certificates; the library does not expose a flag to disable verification.
