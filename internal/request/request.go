package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	before, _, ok := bytes.Cut(data, []byte("\r\n"))
	if !ok {
		return nil, errors.New("Invalid request")
	}

	line, err := parseRequestLine(before)
	if err != nil {
		return nil, err
	}

	request := Request{
		RequestLine: *line,
	}

	return &request, nil
}

func parseRequestLine(line []byte) (*RequestLine, error) {
	s := string(line)

	parts := strings.Split(s, " ")
	if len(parts) != 3 {
		return nil, errors.New("Invalid request")
	}

	err := isAllUppercase(parts[0])
	if err != nil {
		return nil, err
	}

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, errors.New("Invalid request")
	}

	if versionParts[0] != "HTTP" {
		return nil, errors.New("Invalid request")
	}

	if versionParts[1] != "1.1" {
		return nil, errors.New("Invalid request")
	}

	requestLine := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   versionParts[1],
	}

	return &requestLine, nil
}

func isAllUppercase(method string) error {
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return errors.New("invalid method")
		}
	}
	return nil
}
