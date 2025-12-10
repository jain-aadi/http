# Minimal HTTP Server in Go (Built From Scratch Over TCP)

A lightweight HTTP/1.1 server implemented directly over TCP sockets without using Go’s `net/http` package. The goal is to understand the protocol from first principles: request parsing, header handling, connection lifecycle, and streaming responses.

The project also includes a reverse-proxy style endpoint that forwards requests to `httpbin.org` and streams chunked responses back to the client.

---

## Features

### HTTP/1.1 Request Parsing
- Manual parsing of:
  - Request line (method, path, version)
  - Headers (case-insensitive)
  - Message body
- Graceful handling of malformed input with appropriate error responses.

### Persistent Connections
- Basic keep-alive behavior  
- Clean connection teardown  
- Concurrency model using goroutines  

### Streaming Support
- `/httpbin/stream/<N>` forwards to  
  `https://httpbin.org/stream/<N>`
- Response body is streamed back to the client without buffering, preserving upstream chunking behavior.

### Reverse Proxy Behavior
- For paths beginning with `/httpbin/stream`, the server:
  - Forwards the request to upstream
  - Streams upstream bytes as they arrive
  - Preserves chunked transfer semantics

### From-Scratch HTTP Server Core
- Built directly using:
  - `net.Listen`
  - `net.Conn`
  - Manual reading/parsing with `bufio.Reader`
  - Scratch implementation of my own reader and writer for custom usablity

---

## Why This Project Exists

This project explores what usually stays hidden behind high-level web frameworks. It answers questions like:

- What actually arrives over a TCP socket when a browser says “GET /”?
- How does chunked transfer encoding behave under streaming conditions?
- What does an HTTP server need to handle before frameworks take over?

It’s a learning-driven, low-level implementation for building real intuition about networking and protocols.

---

## Usage

### Run the server
```bash
go run .
