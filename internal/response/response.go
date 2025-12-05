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

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

type Writer struct {
	W io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		W: w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
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
	_, err := w.W.Write(statusLine)

	return err
}

func (w *Writer) WriteHeaders(h *headers.Headers) error {
	b := []byte{}

	h.ForEach(func(k, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", k, v)
	})

	b = fmt.Append(b, "\r\n")

	_, err := w.W.Write(b)

	return err
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	n, err := w.W.Write(body)
	return n, err
}
