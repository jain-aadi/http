package server

import (
	"fmt"
	"http_server/internal/request"
	"http_server/internal/response"
	"io"
	"net"
)

type Server struct {
	closed  bool
	handler Handler
}

type HandlerError struct {
	Msg        string
	StatusCode response.StatusCode
}

type Handler func(w *response.Writer, request *request.Request)

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		if s.closed {
			return
		}

		go runConnection(s, conn)
	}
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)

	r, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}

	s.handler(responseWriter, r)
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

func (s *Server) Close() error {
	s.closed = true
	return nil
}
