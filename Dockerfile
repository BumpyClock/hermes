FROM golang:1.24.6-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o parser cmd/parser/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/parser .
COPY --from=builder /app/internal/fixtures ./internal/fixtures

CMD ["./parser"]