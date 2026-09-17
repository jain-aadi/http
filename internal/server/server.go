package server

import (
	"fmt"
	"http_server/internal/request"
	"http_server/internal/response"
	"io"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

type Server struct {
	closed            bool
	handler           Handler
	activeConnections atomic.Int64
	totalConnections  atomic.Int64
	totalRequests     atomic.Int64
	failedRequests    atomic.Int64
	totalHandlerNanos atomic.Int64
}

type Handler func(w *response.Writer, request *request.Request)

// Metrics is a point-in-time snapshot for local diagnostics and demo tooling.
type Metrics struct {
	ActiveConnections int64     `json:"activeConnections"`
	TotalConnections  int64     `json:"totalConnections"`
	TotalRequests     int64     `json:"totalRequests"`
	FailedRequests    int64     `json:"failedRequests"`
	AverageLatencyMS  float64   `json:"averageLatencyMs"`
	Goroutines        int       `json:"goroutines"`
	HeapAllocBytes    uint64    `json:"heapAllocBytes"`
	SystemMemoryBytes uint64    `json:"systemMemoryBytes"`
	GCCount           uint32    `json:"gcCount"`
	CapturedAt        time.Time `json:"capturedAt"`
}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection:", err)
			return
		}

		if s.closed {
			conn.Close()
			return
		}

		s.totalConnections.Add(1)
		s.activeConnections.Add(1)
		go runConnection(s, conn)
	}
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()
	defer s.activeConnections.Add(-1)

	responseWriter := response.NewWriter(conn)

	r, err := request.RequestFromReader(conn)
	if err != nil {
		s.failedRequests.Add(1)
		fmt.Println(err)
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		responseWriter.WriteBody(Respond400())
		return
	}

	s.totalRequests.Add(1)
	started := time.Now()
	s.handler(responseWriter, r)
	s.totalHandlerNanos.Add(time.Since(started).Nanoseconds())
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		closed:  false,
		handler: handler,
	}
	go runServer(server, listener)

	return server, nil
}

// Stats returns lightweight live measurements for the local diagnostics sidecar.
func (s *Server) Stats() Metrics {
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)

	totalRequests := s.totalRequests.Load()
	averageLatencyMS := 0.0
	if totalRequests > 0 {
		averageLatencyMS = float64(s.totalHandlerNanos.Load()) / float64(totalRequests) / float64(time.Millisecond)
	}

	return Metrics{
		ActiveConnections: s.activeConnections.Load(),
		TotalConnections:  s.totalConnections.Load(),
		TotalRequests:     totalRequests,
		FailedRequests:    s.failedRequests.Load(),
		AverageLatencyMS:  averageLatencyMS,
		Goroutines:        runtime.NumGoroutine(),
		HeapAllocBytes:    memory.HeapAlloc,
		SystemMemoryBytes: memory.Sys,
		GCCount:           memory.NumGC,
		CapturedAt:        time.Now(),
	}
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}

func Respond400() []byte {
	return []byte(`<html>
		<head><title>400 Bad Request</title></head>
		<body><h1>400 Bad Request</h1><p>Your request could not be understood by the server due to malformed syntax.</p></body>
		</html>`)
}

func Respond500() []byte {
	return []byte(`<html>
		<head><title>500 Internal Server Error</title></head>
		<body><h1>500 Internal Server Error</h1><p>The server encountered an unexpected condition which prevented it from fulfilling the request.</p></body>
		</html>`)
}

func Respond200() []byte {
	return []byte(`<html>
		<head><title>200 OK</title></head>
		<body><h1>200 OK</h1><p>Your request was successful.</p></body>
		</html>`)
}
