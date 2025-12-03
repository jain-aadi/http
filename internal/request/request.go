package request

import (
	"fmt"
	"http_server/internal/headers"
	"io"
	"strings"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateDone    parserState = "done"
	StateHeaders parserState = "headers"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	State       parserState
	Headers     *headers.Headers
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		switch r.State {

		case StateInit:
			rl, rest, err := ParseRequestLine(string(data[read:]))
			if err != nil {
				return 0, err
			}
			if len(rest) == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += len([]byte(rest))
			r.State = StateHeaders

		case StateDone:
			break outer

		case StateHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}

			read += n
			if done {
				r.State = StateDone
			}

		default:
			panic("unknown state, programming error")
		}
	}

	return read, nil
}

func (r *Request) isDone() bool {
	return r.State == StateDone
}

func newRequest() *Request {
	return &Request{
		State:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var ErrorInvalidRequestLine = fmt.Errorf("invalid request line")

var Seperator = "\r\n"

func ParseRequestLine(line string) (*RequestLine, string, error) {
	idx := strings.Index(line, Seperator)
	if idx == -1 {
		return nil, line, nil
	}

	startLine := line[:idx]
	restOfMsg := line[idx+len(Seperator):]

	lineParts := strings.Split(startLine, " ")
	httpParts := strings.Split(lineParts[2], "/")

	// lineParts should be METHOD, PATH, HTTP protocol
	// httpParts should be HTTP, 1.1
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, restOfMsg, ErrorInvalidRequestLine
	}

	rl := &RequestLine{
		Method:        lineParts[0],
		RequestTarget: lineParts[1],
		HttpVersion:   httpParts[1],
	}

	return rl, restOfMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// a very long header or body will overflow this buffer
	buf := make([]byte, 1024) // 1024 bcoz power of 2 looks good :)
	bufIdx := 0

	for !request.isDone() {
		n, err := reader.Read(buf[bufIdx:]) // read into buf, n is number of bytes read
		if err != nil {
			return nil, fmt.Errorf("unable to read: %w", err)
		}

		bufIdx += n

		readn, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readn:bufIdx]) // copy unread data to beginning of buf
		bufIdx -= readn
	}

	return request, nil

}
