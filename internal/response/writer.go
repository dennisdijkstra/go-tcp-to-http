package response

import (
	"fmt"
	"io"

	"github.com/dennisdijkstra/go-tcp-to-http/internal/headers"
)

type Writer struct {
	writer io.Writer
	state  writerState
}

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  writerStateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateStatusLine {
		return fmt.Errorf("cannot write status line, current state: %d", w.state)
	}
	defer func() { w.state = writerStateHeaders }()

	_, err := w.writer.Write(getStatusLine(statusCode))
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writerStateHeaders {
		return fmt.Errorf("cannot write headers, current state: %d", w.state)
	}
	defer func() { w.state = writerStateBody }()

	for k, v := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body, current state: %d", w.state)
	}

	return w.writer.Write(p)
}
