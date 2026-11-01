.PHONY: build test test-race test-release test-compatibility test-hooks verify lint clean install deps benchmark docker-build

PACKAGES ?= ./...
BENCH ?= .
BENCHFLAGS ?=

build:
	go build -o bin/hermes ./cmd/hermes

test:
	go test -cover $(PACKAGES)

test-race:
	go test -race -coverprofile=coverage.out $(PACKAGES)

test-release:
	PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts -p 'test_*release*.py' -v

test-compatibility:
	PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/compatibility -p 'test_*.py' -v

test-hooks:
	PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts -p 'test_pre_push.py' -v

verify: lint test-race build test-release test-compatibility test-hooks

lint:
	golangci-lint run

clean:
	rm -rf bin/
	go clean -testcache

install:
	go install ./cmd/hermes

deps:
	go mod download
	go mod tidy

benchmark:
	go test -run '^$$' -bench '$(BENCH)' -benchmem $(BENCHFLAGS) $(PACKAGES)

docker-build:
	docker build -t hermes:latest .
