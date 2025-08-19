.PHONY: build test lint clean install deps benchmark run-fixtures

# Go version check
GO_VERSION := 1.24.6
CURRENT_GO := $(shell go version | cut -d' ' -f3 | sed 's/go//')

build:
	go build -o bin/parser cmd/parser/main.go

test:
	go test -v -cover ./...

test-compatibility:
	go test -v ./internal/compatibility/...

lint:
	golangci-lint run

clean:
	rm -rf bin/
	go clean -testcache

install:
	go install ./cmd/parser

deps:
	go mod download
	go mod tidy

benchmark:
	go test -bench=. -benchmem ./...

run-fixtures:
	@echo "Testing with fixtures..."
	go test -v ./pkg/extractors/... -fixtures

copy-fixtures:
	@echo "Copying fixtures from JS project..."
	cp -r ../fixtures/* internal/fixtures/

docker-build:
	docker build -t parser-go:latest .

# Development helpers
dev-setup: deps copy-fixtures
	@echo "Development environment ready"

watch:
	air -c .air.toml