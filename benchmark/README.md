# Parser Benchmark Tool

A cross-platform Node.js script to compare the performance of the Hermes Go parser against the original JavaScript Postlight Parser.

## Features

- **🧹 Clean setup**: Automatically cleans output directories before each run
- **📦 Self-contained**: Installs `@postlight/parser` from npm automatically  
- **⚡ Side-by-side comparison**: Tests both parsers on the same URLs
- **📊 Detailed metrics**: Execution time, file sizes, success rates
- **🔄 Multiple formats**: Tests both JSON and Markdown output formats
- **💾 Output files**: Saves actual parser outputs for manual comparison

## Usage

```bash
cd parser-go/benchmark
node test-comparison.js [urls-file]
```

**Example:**
```bash
node test-comparison.js ./testurls.txt
```

If no URL file is provided, it defaults to `../../testurls.txt`.

## Output

The script creates the following structure:

```
../../test-output/
├── js/
│   ├── json/       # JavaScript parser JSON outputs
│   └── markdown/   # JavaScript parser markdown outputs
├── go/
│   ├── json/       # Go parser JSON outputs  
│   └── markdown/   # Go parser markdown outputs
└── comparison-report.json  # Detailed performance metrics
```

## Requirements

- Node.js (any recent version)
- Go (for building the Hermes parser)
- Internet connection (for npm install)

## Sample Results

```
🎯 Comparison Complete!
=========================
Total time: 10.4s
JSON - JS: 1/1, Go: 1/1
Markdown - JS: 1/1, Go: 1/1

📁 Results saved to: ../../test-output/
📊 Comparison report: ../../test-output/comparison-report.json
```

The comparison report contains detailed metrics for analysis:
- Execution times per parser per format
- Success/failure rates
- Output file sizes
- Error details (if any)