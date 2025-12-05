package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: make(map[string]string),
	}
}

func (h *Headers) ForEach(f func(k, v string)) {
	for k, v := range h.headers {
		f(k, v)
	}
}

var rn = []byte("\r\n")

func parseHeader(line []byte) (string, string, error) {
	parts := bytes.SplitN(line, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid field line")
	}

	key := string(parts[0])
	value := string(bytes.TrimSpace(parts[1]))

	if strings.HasSuffix(key, " ") {
		return "", "", fmt.Errorf("invalid field name (trailing space in key)")
	}

	return key, value, nil

}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Replace(name, value string) {
	n := strings.ToLower(name)

	h.headers[n] = value
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)

	if val, ok := h.headers[name]; ok {
		val += ", " + value
	}

	h.headers[name] = value
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}

		// empty header situation
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}

		name, val, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("invalid header field name")
		}

		read += idx + len(rn)
		h.Set(name, val)
	}

	return read, done, nil
}

func isToken(str []byte) bool {
	for _, b := range str {
		found := false

		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') {
			found = true
		}

		switch b {
		case '!', '#', '$', '%', '&', '*', '+', '-', '.', '^', '_', '`', '|', '~', '\'':
			found = true
		}

		if !found {
			return false
		}
	}

	return true
}
