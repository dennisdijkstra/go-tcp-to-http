package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if h == nil {
		return 0, false, errors.New("nil headers")
	}

	index := bytes.Index(data, []byte(crlf))
	if index == -1 {
		return 0, false, nil
	}

	if index == 0 {
		return len(crlf), true, nil
	}

	headerLine := data[:index]

	parts := strings.SplitN(string(headerLine), ":", 2)
	if len(parts) != 2 {
		return 0, false, errors.New("Invalid header format")
	}

	key := strings.ToLower(parts[0])
	if key != strings.TrimSpace(key) {
		return 0, false, errors.New("Invalid header format")
	}

	for i := 0; i < len(key); i++ {
		char := key[i]
		if !isTokenChar(char) {
			return 0, false, errors.New("Invalid header format")
		}
	}

	value := strings.TrimSpace(parts[1])
	h[key] = value

	return index + len(crlf), false, nil
}

func isTokenChar(b byte) bool {
	return isAlphaNumeric(b) || isSpecialChar(b)
}

func isAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func isSpecialChar(b byte) bool {
	switch b {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}
