# Extraction performance checks

These benchmarks isolate extraction costs from network latency. They do not measure complete URL requests.

## Workloads

| Benchmark | Workload | Timed work |
| --- | --- | --- |
| `BenchmarkMetadataExtraction` | Three metadata fields and 256 article paragraphs | Title, author, and date extraction from one parsed document |
| `BenchmarkGenericAuthorExtractor_Extract` | A small document with author metadata | Author extraction from one parsed document |
| `BenchmarkResponsiveImageLinks` | 32 images with two `srcset` candidates each | Document parse and absolute URL conversion |
| `BenchmarkWithinComment` | An author node with uppercase ancestor attributes | Comment detection through the ancestor chain |

Metadata extractors share the parsed document. They must not modify its nodes during metadata reads.

The response buffer pool retains small byte buffers. Local string builders do not retain storage after `Reset`.

## Comparison procedure

1. Record the revision, Go version, hardware, and dependency versions.
2. Run the same benchmark files on both revisions with other CPU-intensive tasks stopped.
3. Compare latency, bytes per operation, and allocations per operation across repeated samples.
4. Run correctness checks separately from benchmarks.
5. Report sample ranges and distinguish operation-level results from full-request results.

```sh
go version
git rev-parse HEAD
go test ./internal/extractors/generic ./internal/utils/dom \
  -run '^$' \
  -bench 'BenchmarkMetadataExtraction|BenchmarkGenericAuthorExtractor_Extract|BenchmarkResponsiveImageLinks|BenchmarkWithinComment' \
  -benchmem -count=5
```

## Correctness checks

```sh
go test ./internal/extractors/generic ./internal/utils/dom ./internal/pools ./internal/cleaners ./internal/parser
go test -race ./internal/extractors/generic ./internal/utils/dom ./internal/pools ./internal/cleaners ./internal/parser
go vet ./...
```

The existing tests cover metadata priority, fallback selectors, duplicate metadata, URL conversion, `srcset`, and comment ancestry.
