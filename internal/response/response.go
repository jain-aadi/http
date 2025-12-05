package response

import (
	"fmt"
	"http_server/internal/headers"
	"io"
)

type Response struct {
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var statusLine []byte = nil

	switch statusCode {
	case StatusOK:
		statusLine = ([]byte("HTTP/1.1 200 OK"))
	case StatusBadRequest:
		statusLine = ([]byte("HTTP/1.1 400 Bad Request"))
	case StatusInternalServerError:
		statusLine = ([]byte("HTTP/1.1 500 Internal Server Error"))
	default:
		return fmt.Errorf("unrecognized error code")
	}

	statusLine = fmt.Append(statusLine, "\r\n")
	_, err := w.Write(statusLine)

	return err
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, h *headers.Headers) error {
	b := []byte{}

	h.ForEach(func(k, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", k, v)
	})

	b = fmt.Append(b, "\r\n")

	_, err := w.Write(b)

	return err
}
