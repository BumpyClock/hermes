.PHONY: build test lint clean install deps benchmark

build:
	go build -o bin/hermes cmd/hermes/main.go

test:
	go test -cover ./...

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
	go test -bench=. -benchmem ./...

docker-build:
	docker build -t hermes:latest .

watch:
	air -c .air.toml