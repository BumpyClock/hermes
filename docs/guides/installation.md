# Installation

Install the Hermes CLI or add Hermes to a Go project.

## Contents

- [Requirements](#requirements)
- [Installation methods](#installation-methods)
- [First use](#first-use)
- [Environment](#environment)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

## Requirements

- Go version declared in [go.mod](../../go.mod)
- Linux, macOS, or Windows

## Installation methods

### Go install

```bash
go install github.com/BumpyClock/hermes/cmd/hermes@latest
```

Go installs the `hermes` CLI in `GOBIN`, or in the `bin` directory under `GOPATH` when `GOBIN` is empty.

### Build from source

```bash
git clone https://github.com/BumpyClock/hermes.git
cd hermes
make build
make install
```

## First use

Check the CLI version:

```bash
hermes version
```

Parse a URL:

```bash
hermes parse https://example.com/article
```

Add Hermes to your Go module:

```bash
go get github.com/BumpyClock/hermes
```

Use Hermes as a library:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/BumpyClock/hermes"
)

func main() {
    client := hermes.New()
    res, err := client.Parse(context.Background(), "https://example.com")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(res.Title)
}
```

## Environment

Check the Go version and binary directory:

```bash
go version
go env GOPATH
go env GOBIN
```

Add the Go binary directory to `PATH` if your shell cannot find `hermes`.

Configure the library with [client options](../api/configuration.md).

## Verification

Replace the example URLs with article URLs that you can access:

```bash
hermes parse -f markdown https://www.nytimes.com/section/technology
hermes parse --timing https://example.com
hermes parse --concurrency 5 https://example.com/1 https://example.com/2
```

## Troubleshooting

- For timeout errors, increase the limit with `--timeout 60s` or `hermes.WithTimeout(60 * time.Second)`.
- For large batches, reduce concurrent requests with `--concurrency`.
- For TLS certificate errors, check the system CA certificates. The CLI has no flag to disable certificate verification.
