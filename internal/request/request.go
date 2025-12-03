package request

import (
	"fmt"
	"io"
	"strings"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
}

var ErrorInvalidRequestLine = fmt.Errorf("invalid request line")
var ErrorIncompleteStartLine = fmt.Errorf("missing complete request line")

var Seperator = "\r\n"

func ParseRequestLine(line string) (*RequestLine, string, error) {
	idx := strings.Index(line, Seperator)
	if idx == -1 {
		return nil, line, ErrorIncompleteStartLine
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
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to io.ReadAll: %w", err)
	}

	str := string(data)

	rl, _, err := ParseRequestLine(str)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *rl,
	}, err

}
