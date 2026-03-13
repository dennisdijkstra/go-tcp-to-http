# Go TCP to HTTP Project

Go project that builds HTTP handling from raw TCP sockets upward. The codebase includes request parsing, header handling, response writing, and a small HTTP server with custom route behavior (including chunked responses and trailers).

## Tech Stack

- Go 1.25+
- Go standard library
- Testify (tests)

## Project Structure

- `assets/` - static files used by the HTTP server (for example media responses)
- `cmd/` - runnable entrypoints
	- `httpserver/` - custom TCP-backed HTTP server with route handlers
	- `tcplistener/` - low-level TCP request inspection utility
	- `udpsender/` - small UDP sender utility for local experiments
- `internal/` - reusable server and protocol internals
	- `headers/` - HTTP header parsing and validation
	- `request/` - HTTP request parsing from readers
	- `response/` - status line, headers, body/chunked writing helpers
	- `server/` - TCP accept loop and handler orchestration

## Getting Started

1. Install Go 1.25+.
2. Download dependencies:

```bash
go mod download
```

3. Start the HTTP server:

```bash
go run ./cmd/httpserver
```

4. In another terminal, make a request:

```bash
curl -i http://localhost:42069/
```

## Development

- Run tests:

```bash
go test ./...
```
