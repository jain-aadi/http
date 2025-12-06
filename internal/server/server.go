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

type Handler func(w *response.Writer, request *request.Request)

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		responseWriter := response.NewWriter(conn)

		if err != nil {
			fmt.Println("Error accepting connection:", err)
			responseWriter.WriteStatusLine(response.StatusInternalServerError)
			responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
			responseWriter.WriteBody(Respond500())
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
		responseWriter.WriteBody(Respond400())
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
