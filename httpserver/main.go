package main

import (
	"fmt"
	"http_server/internal/request"
	"http_server/internal/response"
	"http_server/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 8000

func main() {
	server, err := server.Serve(port, func(w *response.Writer, r *request.Request) {
		h := response.GetDefaultHeaders(0)
		status := response.StatusOK
		body := respond200()

		if r.RequestLine.RequestTarget == "/yourproblem" {
			status = response.StatusBadRequest
			body = respond400()

		} else if r.RequestLine.RequestTarget == "/myproblem" {
			status = response.StatusInternalServerError
			body = respond500()

		}

		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-Type", "text/html")

		w.WriteStatusLine(status)
		w.WriteHeaders(h)
		w.WriteBody(body)

	})

	if err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}

	defer server.Close()
	fmt.Println("Server started on port:", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nServer gracefully stopped.")
}

func respond400() []byte {
	return []byte(`<html>
		<head><title>400 Bad Request</title></head>
		<body><h1>400 Bad Request</h1><p>Your request could not be understood by the server due to malformed syntax.</p></body>
		</html>`)
}

func respond500() []byte {
	return []byte(`<html>
		<head><title>500 Internal Server Error</title></head>
		<body><h1>500 Internal Server Error</h1><p>The server encountered an unexpected condition which prevented it from fulfilling the request.</p></body>
		</html>`)
}

func respond200() []byte {
	return []byte(`<html>
		<head><title>200 OK</title></head>
		<body><h1>200 OK</h1><p>Your request was successful.</p></body>
		</html>`)
}
