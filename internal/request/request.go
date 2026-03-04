package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine

	state requestState
}

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		state: requestStateInitialized,
	}

	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				req.state = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:readToIndex])
		readToIndex -= numBytesParsed
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, bytesConsumed, err := parseRequestLine(data)

		if err != nil {
			return 0, err
		}

		if bytesConsumed > 0 {
			r.RequestLine = *requestLine
			r.state = requestStateDone
		}

		return bytesConsumed, nil
	case requestStateDone:
		return 0, errors.New("We're done already")
	default:
		return 0, errors.New("An unknown error occured")
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	index := bytes.Index(data, []byte(crlf))
	if index == -1 {
		return nil, 0, nil
	}

	requestLineString := string(data[:index])

	parts := strings.Split(requestLineString, " ")
	if len(parts) != 3 {
		return nil, 0, errors.New("Invalid request")
	}

	method := parts[0]
	err := isAllUppercase(method)
	if err != nil {
		return nil, 0, err
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, 0, errors.New("Invalid request")
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, 0, errors.New("Invalid request")
	}

	version := versionParts[1]
	if version != "1.1" {
		return nil, 0, errors.New("Invalid request")
	}

	noBytesConsumed := index + len(crlf)
	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, noBytesConsumed, nil
}

func isAllUppercase(method string) error {
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return errors.New("invalid method")
		}
	}
	return nil
}
