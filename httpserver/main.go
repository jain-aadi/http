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
		body := server.Respond200()

		fmt.Println("DEBUG: Request received for path:", r.RequestLine.RequestTarget)

		switch r.RequestLine.RequestTarget {
		case "/yourproblem":
			status = response.StatusBadRequest
			body = server.Respond400()

		case "/myproblem":
			status = response.StatusInternalServerError
			body = server.Respond500()

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
