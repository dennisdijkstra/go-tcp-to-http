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
	writerStateTrailers
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

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.state)
	}

	chunkSize := len(p)
	nHeader, err := fmt.Fprintf(w.writer, "%x\r\n", chunkSize)
	if err != nil {
		return 0, err
	}

	nData, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}

	nTrailer, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}

	nTotal := nHeader + nData + nTrailer
	return nTotal, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.state)
	}

	n, err := w.writer.Write([]byte("0\r\n"))
	if err != nil {
		return 0, err
	}

	w.state = writerStateTrailers
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writerStateTrailers {
		return fmt.Errorf("cannot write trailers in state %d", w.state)
	}

	for k, v := range h {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}
