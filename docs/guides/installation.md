# Installation & Setup Guide

This guide covers everything you need to install and set up Hermes for development and production use.

## Table of Contents

- [System Requirements](#system-requirements)
- [Installation Methods](#installation-methods)
- [Quick Start](#quick-start)
- [Environment Setup](#environment-setup)
- [Verification](#verification)
- [Docker Installation](#docker-installation)
- [Troubleshooting](#troubleshooting)

## System Requirements

### Minimum Requirements

- **Go**: Version 1.24.6 or later
- **Operating System**: Linux, macOS, or Windows
- **Memory**: 512 MB RAM minimum, 2 GB recommended
- **Disk Space**: 100 MB for installation

### Recommended Requirements

- **Go**: Latest stable version (1.24.6+)
- **Memory**: 4 GB RAM for production workloads
- **CPU**: Multi-core processor for concurrent processing
- **Network**: Stable internet connection for web scraping

### Dependencies

Hermes uses carefully selected dependencies for optimal performance:

- **goquery** (v1.10.3): jQuery-like DOM manipulation
- **html-to-markdown** (v1.6.0): HTML to Markdown conversion
- **go-dateparser** (v1.2.4): Flexible date parsing
- **bluemonday** (v1.0.27): HTML sanitization
- **chardet** (v0.0.0-20230101081208): Character encoding detection
- **cobra** (v1.9.1): CLI framework

## Installation Methods

### Method 1: Go Install (Recommended)

Install the latest release directly from the repository:

```bash
go install github.com/BumpyClock/hermes/cmd/parser@latest
```

This installs the `parser` CLI tool to your `$GOPATH/bin` directory.

### Method 2: Build from Source

Clone and build from source for development or customization:

```bash
# Clone the repository
git clone https://github.com/BumpyClock/hermes.git
cd hermes

# Build the binary
make build

# Install to $GOPATH/bin
make install
```

### Method 3: Download Pre-built Binary

Download pre-built binaries from the releases page:

```bash
# For Linux x86_64
curl -L -o parser https://github.com/BumpyClock/hermes/releases/latest/download/parser-linux-amd64
chmod +x parser
sudo mv parser /usr/local/bin/

# For macOS
curl -L -o parser https://github.com/BumpyClock/hermes/releases/latest/download/parser-darwin-amd64
chmod +x parser
sudo mv parser /usr/local/bin/

# For Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/BumpyClock/hermes/releases/latest/download/parser-windows-amd64.exe" -OutFile "parser.exe"
```

### Method 4: Package Managers

#### Homebrew (macOS)

```bash
brew tap BumpyClock/hermes
brew install hermes
```

#### Snap (Linux)

```bash
sudo snap install hermes
```

#### Chocolatey (Windows)

```bash
choco install hermes
```

## Quick Start

### 1. Verify Installation

```bash
parser version
```

Expected output:
```
Hermes v0.1.0
Go version: 1.24.6
```

### 2. Basic Usage

Parse a single URL:

```bash
parser parse https://example.com/article
```

Parse with Markdown output:

```bash
parser parse -f markdown https://example.com/article
```

Save to file:

```bash
parser parse -f markdown -o article.md https://example.com/article
```

### 3. Library Usage

Create a simple Go program:

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    // Create parser
    p := parser.New()
    
    // Parse a URL
    result, err := p.Parse("https://example.com/article", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    if result.IsError() {
        log.Fatal(result.Message)
    }
    
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Author: %s\n", result.Author)
    fmt.Printf("Word Count: %d\n", result.WordCount)
}
```

Run the program:

```bash
go mod init myproject
go get github.com/BumpyClock/hermes
go run main.go
```

## Environment Setup

### 1. Go Environment

Ensure Go is properly configured:

```bash
# Check Go version
go version

# Check Go environment
go env GOPATH
go env GOROOT

# Ensure $GOPATH/bin is in your PATH
echo $PATH | grep -q "$(go env GOPATH)/bin" || echo "Add $(go env GOPATH)/bin to your PATH"
```

### 2. Environment Variables

Set optional environment variables for configuration:

```bash
# Parser configuration
export HERMES_CONTENT_TYPE=markdown
export HERMES_FETCH_ALL_PAGES=true
export HERMES_FALLBACK=true

# HTTP configuration
export HERMES_TIMEOUT=30s
export HERMES_MAX_REDIRECTS=10
export HERMES_USER_AGENT="Hermes/1.0 (+https://yoursite.com/bot)"

# Performance configuration
export HERMES_MAX_CONCURRENCY=100
export HERMES_POOL_SIZE=1000
export HERMES_WORKER_COUNT=8

# Security configuration
export HERMES_VALIDATE_SSL=true
export HERMES_MAX_URL_LENGTH=2048
```

Add to your shell profile (`.bashrc`, `.zshrc`, etc.):

```bash
echo 'export HERMES_CONTENT_TYPE=markdown' >> ~/.bashrc
source ~/.bashrc
```

### 3. Configuration File

Create a configuration file for persistent settings:

```bash
mkdir -p ~/.config/hermes
```

Create `~/.config/hermes/config.yaml`:

```yaml
# Hermes Configuration
parser:
  content_type: markdown
  fetch_all_pages: true
  fallback: true
  
http:
  timeout: 30s
  max_redirects: 10
  user_agent: "Hermes/1.0 (+https://yoursite.com/bot)"
  
performance:
  max_concurrency: 100
  pool_size: 1000
  worker_count: 8
  
security:
  validate_ssl: true
  max_url_length: 2048
```

## Verification

### 1. CLI Verification

Test CLI functionality:

```bash
# Basic parsing
parser parse https://www.theguardian.com/technology

# Format options
parser parse -f markdown https://www.nytimes.com/section/technology

# Multiple URLs
parser parse https://example.com/1 https://example.com/2 https://example.com/3

# Custom headers
parser parse --headers '{"User-Agent": "MyBot/1.0"}' https://example.com

# Timing information
parser parse --timing https://example.com
```

### 2. Library Verification

Test library functionality:

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    // Test basic parsing
    testBasicParsing()
    
    // Test configuration
    testConfiguration()
    
    // Test error handling
    testErrorHandling()
    
    // Test performance
    testPerformance()
}

func testBasicParsing() {
    fmt.Println("Testing basic parsing...")
    
    p := parser.New()
    result, err := p.Parse("https://www.theguardian.com/technology", nil)
    if err != nil {
        log.Printf("Basic parsing failed: %v", err)
        return
    }
    
    if result.IsError() {
        log.Printf("Extraction failed: %s", result.Message)
        return
    }
    
    fmt.Printf("✓ Title: %s\n", result.Title)
    fmt.Printf("✓ Word count: %d\n", result.WordCount)
    fmt.Printf("✓ Domain: %s\n", result.Domain)
}

func testConfiguration() {
    fmt.Println("Testing configuration...")
    
    opts := &parser.ParserOptions{
        ContentType: "markdown",
        FetchAllPages: false,
        Headers: map[string]string{
            "User-Agent": "Hermes-Test/1.0",
        },
    }
    
    p := parser.New(opts)
    result, err := p.Parse("https://example.com", opts)
    if err != nil {
        log.Printf("Configuration test failed: %v", err)
        return
    }
    
    fmt.Printf("✓ Configuration test passed\n")
}

func testErrorHandling() {
    fmt.Println("Testing error handling...")
    
    p := parser.New()
    
    // Test invalid URL
    _, err := p.Parse("invalid-url", nil)
    if err != nil {
        fmt.Printf("✓ Invalid URL properly handled: %v\n", err)
    }
    
    // Test non-existent domain
    result, err := p.Parse("https://this-domain-does-not-exist-12345.com", nil)
    if err != nil {
        fmt.Printf("✓ Network error properly handled: %v\n", err)
    } else if result.IsError() {
        fmt.Printf("✓ Extraction error properly handled: %s\n", result.Message)
    }
}

func testPerformance() {
    fmt.Println("Testing performance...")
    
    p := parser.New()
    start := time.Now()
    
    result, err := p.Parse("https://www.nytimes.com/section/technology", nil)
    duration := time.Since(start)
    
    if err == nil && !result.IsError() {
        fmt.Printf("✓ Performance test passed: %v\n", duration)
        if duration < 5*time.Second {
            fmt.Printf("✓ Good performance: under 5 seconds\n")
        }
    }
}
```

### 3. Health Check Script

Create a health check script:

```bash
#!/bin/bash
# health-check.sh

echo "Hermes Health Check"
echo "==================="

# Check if parser is installed
if ! command -v parser &> /dev/null; then
    echo "❌ Parser CLI not found"
    exit 1
fi

echo "✓ Parser CLI found"

# Check version
VERSION=$(parser version 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "✓ Version: $VERSION"
else
    echo "❌ Version check failed"
    exit 1
fi

# Test basic functionality
echo "Testing basic functionality..."
RESULT=$(parser parse https://httpbin.org/html 2>/dev/null)
if [ $? -eq 0 ] && [ ! -z "$RESULT" ]; then
    echo "✓ Basic parsing works"
else
    echo "❌ Basic parsing failed"
    exit 1
fi

echo "✓ All health checks passed"
```

Make it executable and run:

```bash
chmod +x health-check.sh
./health-check.sh
```

## Docker Installation

Note: Official pre-built images are not yet published. Use the build-from-source method below.

### 1. Using Pre-built Image

```bash
# Pull the image
docker pull ghcr.io/bumpyclock/hermes:latest

# Run parser
docker run --rm ghcr.io/bumpyclock/hermes:latest parser parse https://example.com

# Run with volume for output
docker run --rm -v $(pwd):/output ghcr.io/bumpyclock/hermes:latest \
    parser parse -f markdown -o /output/article.md https://example.com
```

### 2. Building from Source

```bash
# Clone repository
git clone https://github.com/BumpyClock/hermes.git
cd hermes

# Build Docker image
docker build -t hermes:local .

# Run the image
docker run --rm hermes:local parser parse https://example.com
```

### 3. Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  hermes:
    image: ghcr.io/bumpyclock/hermes:latest
    volumes:
      - ./output:/output
      - ./config:/config
    environment:
      - HERMES_CONTENT_TYPE=markdown
      - HERMES_TIMEOUT=30s
    command: >
      parser parse 
      -f markdown 
      -o /output/articles.json
      https://example.com/article1
      https://example.com/article2
```

Run with:

```bash
docker-compose up
```

### 4. Kubernetes Deployment

Create `hermes-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hermes-parser
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hermes-parser
  template:
    metadata:
      labels:
        app: hermes-parser
    spec:
      containers:
      - name: hermes
        image: ghcr.io/bumpyclock/hermes:latest
        env:
        - name: HERMES_CONTENT_TYPE
          value: "markdown"
        - name: HERMES_MAX_CONCURRENCY
          value: "50"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

Deploy:

```bash
kubectl apply -f hermes-deployment.yaml
```

## Troubleshooting

### Common Issues

#### 1. Command Not Found

**Error:** `command not found: parser`

**Solution:**
```bash
# Check if $GOPATH/bin is in PATH
echo $PATH | grep "$(go env GOPATH)/bin"

# If not, add it
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc

# Or install to /usr/local/bin
sudo cp $(go env GOPATH)/bin/parser /usr/local/bin/
```

#### 2. Permission Denied

**Error:** `permission denied: parser`

**Solution:**
```bash
# Make executable
chmod +x parser

# Or if in /usr/local/bin
sudo chmod +x /usr/local/bin/parser
```

#### 3. Network Timeouts

**Error:** `context deadline exceeded`

**Solution:**
```bash
# Increase timeout in your HTTP client wrapper
export HERMES_TIMEOUT=60s
```

#### 4. Memory Issues

**Error:** `fatal error: runtime: out of memory`

**Solution:**
```bash
# Reduce concurrency
export HERMES_MAX_CONCURRENCY=10

# Increase available memory
export GOGC=100

# Use smaller pool size
export HERMES_POOL_SIZE=100
```

#### 5. SSL Certificate Issues

**Error:** `x509: certificate signed by unknown authority`

**Solution:**
```bash
# Disable SSL validation (not recommended for production)
export HERMES_VALIDATE_SSL=false

# Or update certificates
# On Ubuntu/Debian:
sudo apt-get update && sudo apt-get install ca-certificates

# On macOS:
brew install ca-certificates
```

### Debugging

Enable additional logging in your application by instrumenting around parser calls. The CLI currently does not support a `--verbose` flag.

### Performance Tuning

Optimize for your environment:

```bash
# For high-concurrency scenarios
export HERMES_MAX_CONCURRENCY=200
export HERMES_WORKER_COUNT=16
export HERMES_POOL_SIZE=2000

# For memory-constrained environments
export HERMES_MAX_CONCURRENCY=10
export HERMES_WORKER_COUNT=4
export HERMES_POOL_SIZE=100
export GOGC=50
```

### Getting Help

If you encounter issues not covered here:

1. Check the [GitHub Issues](https://github.com/BumpyClock/hermes/issues)
2. Search for similar problems in closed issues
3. Create a new issue with:
   - Go version (`go version`)
   - Operating system
   - Hermes version (`parser version`)
   - Complete error message
   - Steps to reproduce

## Next Steps

After successful installation:

1. Read the [Basic Usage Guide](basic-usage.md)
2. Explore [API Documentation](../api/parser.md)
3. Check out [Examples](../examples/basic.md)
4. Learn about [Custom Extractors](custom-extractors.md)

For development:

1. Set up [Development Environment](../development/setup.md)
2. Read [Contributing Guidelines](../development/contributing.md)
3. Explore [Architecture Documentation](../architecture/overview.md)
