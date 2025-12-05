package server

import (
	"bytes"
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

type Handler func(w io.Writer, request *request.Request) *HandlerError

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

	headers := response.GetDefaultHeaders(0)

	r, err := request.RequestFromReader(conn)
	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	var body []byte = nil
	var status response.StatusCode = response.StatusOK

	handlerErr := s.handler(writer, r)
	if handlerErr != nil {
		status = handlerErr.StatusCode
		body = []byte(handlerErr.Msg)
	} else {
		body = writer.Bytes()
	}

	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))

	response.WriteStatusLine(conn, status)
	response.WriteHeaders(conn, headers)

	conn.Write(body)
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
