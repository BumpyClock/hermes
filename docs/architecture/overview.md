# Architecture Overview

This overview describes how Hermes is organized internally and how a parse request flows through the system.

## Components

- `hermes.Client`: Public entry point. Manages HTTP client configuration and orchestrates parsing.
- `internal/resource`: Fetching, encoding detection, and DOM preparation.
- `internal/extractors`: Extractor selection (custom vs generic) and field extraction.
- `internal/cleaners`: Field-specific cleaning and normalization.
- `internal/parser`: Orchestration of extraction and result assembly.
- `internal/validation`: URL validation and SSRF protection.
- CLI (`cmd/hermes`): Thin wrapper around the client with concurrency and formatting options.

## System Diagram

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
    I --> J[internal/parser/result]
    J --> K[hermes.Result]
```

## Parse Flow

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
    Parser->>Valid: ValidateURL(ctx, url)
    Valid-->>Parser: ok or error
    Parser->>Res: Fetch + Normalize HTML
    Res-->>Parser: Document
    Parser->>Ext: Select and run extractor
    Ext->>Clean: Clean/normalize fields
    Clean-->>Parser: Clean fields
    Parser-->>Client: Result or ParseError
    Client-->>App: Result or error
```

## Notes on Behavior

- Concurrency: The client is safe to share; use your own goroutines to parallelize parsing.
- Formats: Content format (HTML/Markdown/Text) is selected via `WithContentType`.
- Security: Private network access is blocked by default; enable only when necessary using `WithAllowPrivateNetworks`.

