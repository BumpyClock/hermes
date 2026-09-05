# Parser benchmark

`test-comparison.js` compares the Hermes CLI with the JavaScript Postlight Parser on the same URLs. It tests JSON and Markdown output.

## Requirements

- Node.js and npm.
- The Go version in [`go.mod`](../go.mod).
- Make, if the repository has a `Makefile`.
- Network access for dependencies and test URLs.

## Run the comparison

The script deletes `benchmark/test-output/` before each run. Preserve any results that you need before you run it.

Use only trusted URL files. The script inserts URLs into shell commands without shell-safe argument separation.

From the repository root, run:

```bash
cd benchmark
node test-comparison.js ./testurls.txt
```

The URL file defaults to `./testurls.txt`. The script accepts one `https://` URL per line and ignores other lines.

The script runs `npm install` in the current directory. It then builds `../bin/hermes` with `make build`.
If the parent directory has no `Makefile`, it uses `go build -o bin/hermes ./cmd/hermes` instead.

## Output

The script writes results relative to the `benchmark` directory:

```text
test-output/
├── js/
│   ├── json/
│   └── markdown/
├── go/
│   ├── json/
│   └── markdown/
└── comparison-report.json
```

The report records execution time in milliseconds, success and failure counts, output file sizes in bytes, and errors.
The output files contain parser results for manual comparison.

## Interpret the results

These results measure live requests, not isolated extraction speed. Network conditions and remote page changes affect the comparison.

Each Go measurement includes CLI process startup. The JavaScript parser runs in the benchmark process, and its first measurement includes module load time.

`averageTime` divides total attempt time by the number of successful attempts. Failed attempts contribute to the total, so this is not a success-only average.

For Go benchmarks with allocation metrics, run `make benchmark` from the repository root.
