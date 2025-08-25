# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-08-24

### 🎉 Major Release: Go Module Refactoring Complete

This release represents a complete architectural overhaul of Hermes, transforming it from a CLI tool with internal packages to a proper Go library with a clean public API.

### ✨ Added

#### New Public API
- **Root Package API**: Clean `github.com/BumpyClock/hermes` import instead of internal packages
- **Client Type**: Thread-safe, reusable `Client` struct with connection pooling
- **Functional Options**: Modern Go patterns with `WithTimeout()`, `WithUserAgent()`, etc.
- **Context Support**: Full context.Context integration for cancellation and timeouts
- **Structured Errors**: New `ParseError` type with error codes (`ErrInvalidURL`, `ErrFetch`, etc.)
- **Content Type Options**: `WithContentType()` affects parser extraction, not just output formatting
- **SSRF Protection**: `WithAllowPrivateNetworks()` option for security-conscious deployments

#### Enhanced CLI
- **Batch Processing**: Concurrent URL processing with configurable limits
- **Semaphore Pattern**: Efficient resource management for high-throughput scenarios  
- **Progress Reporting**: Detailed timing metrics with `--timing` flag
- **Graceful Errors**: Partial failure handling - continues processing remaining URLs
- **All Output Formats**: JSON, HTML, Markdown, and Text extraction working correctly

#### Developer Experience
- **Comprehensive Examples**: Basic usage, custom HTTP clients, error handling patterns
- **Migration Guide**: Clear upgrade path from v0.x internal API
- **Testable API**: `Parser` interface for easy mocking in tests
- **Thread Safety**: Client can be safely shared across goroutines

### 🔄 Changed

#### Breaking Changes
- **Package Import**: Use `github.com/BumpyClock/hermes` instead of `/pkg/parser`
- **API Signatures**: All parse methods now require `context.Context` as first parameter
- **Configuration**: Options moved from struct fields to functional options pattern
- **Error Types**: Errors now return structured `*ParseError` instead of generic errors
- **HTTP Client**: Client manages its own HTTP client instead of global singleton

#### Internal Architecture  
- **No Global State**: Removed all global HTTP clients and singletons
- **Context Threading**: Context properly propagated through entire call chain
- **Clean Separation**: Library handles parsing, CLI handles orchestration
- **Memory Efficient**: Removed object pools and streaming complexity (no performance regression)

### 🗑️ Removed

#### Deprecated Features
- **Batch API**: Removed `BatchParse()` - use CLI or implement your own concurrency
- **Worker Pools**: Removed internal worker pool complexity
- **Object Pools**: Removed object pooling (memory usage unchanged)
- **Streaming API**: Removed streaming interface (was unused)
- **Global HTTP Client**: No more global singletons

#### Internal Packages
- **pkg/* Structure**: Moved all packages to `internal/` to hide implementation
- **Public Internal API**: No more accidental dependency on internal types

### 🐛 Fixed

#### Critical Fixes
- **HTTP Client Injection**: HTTP client now properly passed through entire stack
- **Context Cancellation**: Context cancellation works throughout call chain  
- **SSRF Protection**: Private network blocking works correctly with toggle
- **Content Type Extraction**: Markdown/Text formats now extract from parser, not just client formatting
- **Memory Leaks**: Proper resource cleanup and connection pooling
- **Race Conditions**: Thread-safe client design eliminates data races

#### CLI Fixes
- **Format Flag**: `-f markdown` now affects parser extraction, not just output formatting
- **Error Reporting**: Better error messages with timing information
- **Concurrent Safety**: Proper semaphore usage prevents resource exhaustion

### 📈 Performance

- **Memory Usage**: Maintained ~1.6MB memory usage despite architectural changes
- **Connection Reuse**: HTTP client connection pooling for better throughput
- **Concurrent Processing**: CLI supports configurable concurrent URL processing
- **No Regressions**: All performance benchmarks maintained or improved

### 🔧 Internal Improvements

#### Code Quality
- **DRY/KISS Principles**: Eliminated code duplication, simplified architecture
- **Error Handling**: Consistent error propagation and classification
- **Resource Management**: Proper HTTP client lifecycle management
- **Test Coverage**: Comprehensive context cancellation and error handling tests

#### Architecture
- **Layered Design**: Clear separation between public API, internal parser, and resource layers
- **Dependency Injection**: HTTP client and options properly injected throughout stack
- **Security**: Built-in SSRF protection with configurable private network access

### 📚 Documentation

- **Updated README**: Complete rewrite with new API examples
- **Migration Guide**: Clear upgrade instructions from v0.x to v1.0
- **Error Handling**: Comprehensive error handling patterns
- **Examples**: Basic usage, custom clients, concurrent processing patterns

### 🧪 Testing

- **Context Tests**: Comprehensive context cancellation and timeout testing
- **Integration Tests**: Real URL testing with various configurations
- **Error Path Testing**: All error codes and edge cases covered
- **Concurrent Testing**: Thread safety and race condition testing

---

## [0.x] - Legacy Versions

Previous versions used internal package structure and are no longer supported. 
See migration guide in README.md for upgrade instructions.

### Migration Required

If you're using any version before 1.0.0, you **must** update your code to use the new API. 
The old `/pkg/parser` import path no longer works and will cause compilation errors.

**Quick Migration:**

```go
// Old (v0.x)
import "github.com/BumpyClock/hermes/pkg/parser"
p := parser.New()
result, err := p.Parse(url, options)

// New (v1.0+)  
import "github.com/BumpyClock/hermes"
client := hermes.New(hermes.WithTimeout(30*time.Second))
result, err := client.Parse(context.Background(), url)
```

### Support

- **v1.0+**: ✅ Active development and support
- **v0.x**: ❌ No longer supported, please upgrade