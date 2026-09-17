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

### Live Runtime Profiling
- A loopback-only Go `pprof` sidecar runs at `http://127.0.0.1:6060/debug/pprof/`.
- It is separate from the custom raw-TCP server, which continues serving requests on port `8000`.
- The profiler is intended for local development and demos; it is not exposed to the network.
- A live dashboard is available at `http://127.0.0.1:6060/debug/dashboard`. It charts request throughput, active connections, goroutines, average handler latency, heap allocation, garbage collection, and request failures.
- Raw point-in-time metrics are also available as JSON at `http://127.0.0.1:6060/debug/metrics`.
- The dashboard includes a bounded local load runner: choose up to 10,000 requests and 500 concurrent requests, then run the test directly from the page.

---

## Why This Project Exists

This project explores what usually stays hidden behind high-level web frameworks. It answers questions like:

- What actually arrives over a TCP socket when a browser says “GET /”?
- How does chunked transfer encoding behave under streaming conditions?
- What does an HTTP server need to handle before frameworks take over?
- What are different http protocols?
- What is difference between TCP and UDP?

It’s a learning-driven, low-level implementation for building real intuition about networking and protocols.

---

## Load Demo with a Live Goroutine Graph

Start the server:

```bash
go run ./httpserver
```

In a second terminal, start the interactive goroutine visualizer:

```bash
go tool pprof -http=:8081 http://127.0.0.1:6060/debug/pprof/goroutine
```

Open `http://127.0.0.1:6060/debug/dashboard` in a browser. It samples and charts the server every second. Keep the pprof window open as well; at `http://127.0.0.1:8081`, select **Graph** for the current goroutine call graph.

In a third terminal running PowerShell 7, generate continuous concurrent traffic (stop with `Ctrl+C`):

```bash
while ($true) {
  1..200 | ForEach-Object -Parallel {
    Invoke-WebRequest http://127.0.0.1:8000/ -UseBasicParsing | Out-Null
  } -ThrottleLimit 50
}
```

For the recording, show the dashboard building a throughput line, then switch to pprof and refresh the graph while traffic is in flight. Increase the request count and throttle limit gradually to find a level your machine handles comfortably. This is a local illustrative load check—not a production capacity benchmark.

If pprof reports that it cannot execute `dot`, install Graphviz and restart the terminal before re-running pprof:

```bash
winget install graphviz
```

---

