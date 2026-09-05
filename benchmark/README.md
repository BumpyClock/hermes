# Parser comparison

This script compares live extraction output from Postlight Parser and the Hermes CLI.
It does not isolate parser performance.
JavaScript runs in-process, but each Hermes attempt starts a new process.
Each parser also makes a separate network request.

The live comparison requires Node.js, Go, and network access.
It saves JSON and Markdown output under `test-output/js` and `test-output/go`.
The report is `test-output/comparison-report.json`.

## Local checks

Run the regression checks without network access or third-party packages:

```sh
cd benchmark
npm test
```

Run the live comparison from this directory:

```sh
node test-comparison.js testurls.txt
```

The live script installs dependencies, builds the CLI, and replaces the local `test-output` directory.
The URL file accepts one HTTPS URL per line.
The Go process receives each URL as a literal argument, not shell syntax.

## Report metrics

`totalTime` includes successful and failed attempts.
`averageTime` equals `totalTime / totalUrls`, rounded to milliseconds.
An empty input has an average of zero.
Success and failure counts remain separate fields.

These metrics describe attempt latency, not successful-request latency or comparative parser throughput.
Use identical local HTML fixtures and equivalent process boundaries for performance comparisons.

The CLI returns an error when every URL fails, including with `--timing`.
It omits the successful-parse average when there are no successful results.
