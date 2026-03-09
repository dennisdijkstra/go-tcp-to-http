package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dennisdijkstra/go-tcp-to-http/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte

	state          requestState
	bodyLengthRead int
}

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
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
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}

	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d", req.state)
				}
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
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}

		if n == 0 {
			break
		}

		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, bytesConsumed, err := parseRequestLine(data)

		if err != nil {
			return 0, err
		}

		if bytesConsumed > 0 {
			r.RequestLine = *requestLine
			r.state = requestStateParsingHeaders
		}

		return bytesConsumed, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		value, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.state = requestStateDone
			return 0, nil
		}

		contentLength, err := strconv.Atoi(value)
		if err != nil {
			return 0, errors.New("invalid Content-Length header")
		}

		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)

		if r.bodyLengthRead > contentLength {
			return len(data), errors.New("read more body data than specified in Content-Length")
		}

		if r.bodyLengthRead == contentLength {
			r.state = requestStateDone
		}

		return len(data), nil
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
