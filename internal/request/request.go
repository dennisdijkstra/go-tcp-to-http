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
	rawBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestLine, err := parseRequestLine(rawBytes)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

func parseRequestLine(data []byte) (*RequestLine, error) {
	index := bytes.Index(data, []byte("\r\n"))
	if index == -1 {
		return nil, errors.New("Invalid request")
	}

	requestLineString := string(data[:index])

	parts := strings.Split(requestLineString, " ")
	if len(parts) != 3 {
		return nil, errors.New("Invalid request")
	}

	method := parts[0]
	err := isAllUppercase(method)
	if err != nil {
		return nil, err
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, errors.New("Invalid request")
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, errors.New("Invalid request")
	}

	version := versionParts[1]
	if version != "1.1" {
		return nil, errors.New("Invalid request")
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, nil
}

func isAllUppercase(method string) error {
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return errors.New("invalid method")
		}
	}
	return nil
}
