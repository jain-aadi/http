package main

import (
	"fmt"
	"http_server/internal/request"
	"http_server/internal/response"
	"http_server/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 8000

func main() {
	server, err := server.Serve(port, func(w *response.Writer, r *request.Request) {
		h := response.GetDefaultHeaders(0)
		status := response.StatusOK
		body := server.Respond200()

		errFlag := false

		fmt.Println("DEBUG: Request received for path:", r.RequestLine.RequestTarget)

		switch r.RequestLine.RequestTarget {
		case "/yourproblem":
			status = response.StatusBadRequest
			body = server.Respond400()
			errFlag = true

		case "/myproblem":
			status = response.StatusInternalServerError
			body = server.Respond500()
			errFlag = true

		}

		if !errFlag && strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/stream") {
			target := r.RequestLine.RequestTarget

			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
			if err != nil {
				status = response.StatusInternalServerError
				body = server.Respond500()
			} else {
				w.WriteStatusLine(response.StatusOK)
				h.Delete("Content-Length")
				h.Replace("Content-Type", "text/plain")
				h.Set("Transfer-Encoding", "chunked")

				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if err != nil {
						break
					}
					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}

				w.WriteBody([]byte("0\r\n\r\n"))
				return
			}

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
