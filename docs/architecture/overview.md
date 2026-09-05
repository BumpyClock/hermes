# Architecture overview

Hermes separates its public client, resource access, extraction, and result conversion into these components.

## Components

| Component | Role |
| --- | --- |
| `hermes.Client` | Provides the public API and HTTP client configuration. |
| `internal/resource` | Fetches HTML, detects character encoding, and prepares the DOM. |
| `internal/extractors` | Provides custom and generic field extractors. |
| `internal/cleaners` | Cleans and normalizes extracted fields. |
| `internal/parser` | Selects extractors, coordinates extraction, and assembles results. |
| `internal/validation` | Validates URLs against scheme, host, and network rules. |
| `cmd/hermes` | Provides the CLI with concurrency and output format options. |

## Component relationships

```mermaid
flowchart TD
    subgraph Public API
        A[hermes.Client]
    end

    A -->|Parse / ParseHTML| B[internal/parser]
    B --> C[internal/validation]
    B --> D[internal/resource]
    D --> E[HTML Document]
    B --> F{Extractor Selection}
    F -->|Custom| G[internal/extractors/custom]
    F -->|Generic| H[internal/extractors/generic]
    G --> I[internal/cleaners]
    H --> I[internal/cleaners]
    I --> J[internal/parser.Result]
    J --> K[hermes.Result]
```

## Parse request

```mermaid
sequenceDiagram
    participant App as Your App
    participant Client as hermes.Client
    participant Parser as internal/parser
    participant Valid as internal/validation
    participant Res as internal/resource
    participant Ext as internal/extractors
    participant Clean as internal/cleaners

    App->>Client: Parse(ctx, url)
    Client->>Parser: ParseWithContext(ctx, url, opts)
    Parser->>Valid: ValidateParsedURL(ctx, parsedURL, url, opts)
    Valid-->>Parser: ok or error
    Parser->>Res: Fetch + Normalize HTML
    Res-->>Parser: Document
    Parser->>Ext: Select and run extractor
    Ext->>Clean: Clean/normalize fields
    Clean-->>Parser: Clean fields
    Parser-->>Client: Result or ParseError
    Client-->>App: Result or error
```

## Client behavior

The client supports concurrent requests. Reuse a client across goroutines to share its HTTP connection pool.

`WithContentType` selects HTML, Markdown, or text for the extracted content.

URL validation rejects private network addresses by default. Enable `WithAllowPrivateNetworks` only in trusted environments that require private network access.

See [security limitations](../api/configuration.md#security) for redirect and DNS constraints.

`ParseHTML` accepts HTML from the caller instead of a URL fetch. It still validates the supplied URL before extraction.
