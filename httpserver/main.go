package main

import (
	"fmt"
	"http_server/internal/request"
	"http_server/internal/response"
	"http_server/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 8000

func main() {
	server, err := server.Serve(port, func(w io.Writer, r *request.Request) *server.HandlerError {

		if r.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Msg:        "You just made a bad request!\n",
			}

		} else if r.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Msg:        "My Bad!\n",
			}
		} else {
			w.Write([]byte("All Good!\n"))
		}

		return nil
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
