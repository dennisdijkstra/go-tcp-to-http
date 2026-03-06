# Go TCP to HTTP Project

Small Go project for parsing HTTP/1.1 request data from a raw TCP connection.

## Tech Stack

- Go
- Standard library (`net`, `io`, `strings`, `bytes`)
- Testify (for tests)

## Project Structure

- `cmd/tcplistener/main.go` - TCP server that accepts connections and parses request lines
- `cmd/udpsender/main.go` - small UDP sender utility for local networking experiments
- `internal/request/request.go` - incremental HTTP request-line parser
- `internal/headers/headers.go` - HTTP header parser and validator
- `internal/request/request_test.go` - request parser tests
- `internal/headers/headers_test.go` - header parser tests

## Getting Started

1. Install Go 1.25+.
2. Download dependencies:

```bash
go mod download
```

3. Start the TCP listener:

```bash
go run ./cmd/tcplistener
```

4. In another terminal, send a raw HTTP request:

```bash
printf "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\n\r\n" | nc localhost 42069
```

## Development

- Run tests:

```bash
go test ./...
```

- Request parser coverage:
- Method (must be uppercase)
- Request target (path)
- HTTP version (currently `HTTP/1.1`)

- Header parser behavior: token-style header name validation with trimmed values.
