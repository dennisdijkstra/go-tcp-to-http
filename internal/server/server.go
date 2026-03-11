package server

import (
	"bytes"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/dennisdijkstra/go-tcp-to-http/internal/request"
	"github.com/dennisdijkstra/go-tcp-to-http/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func WriteHandlerError(w io.Writer, err *HandlerError) {
	if err == nil {
		return
	}

	if writeErr := response.WriteStatusLine(w, err.StatusCode); writeErr != nil {
		log.Printf("Error writing status line: %v", writeErr)
		return
	}

	headers := response.GetDefaultHeaders(len(err.Message))
	if writeErr := response.WriteHeaders(w, headers); writeErr != nil {
		log.Printf("Error writing headers: %v", writeErr)
		return
	}

	if _, writeErr := w.Write([]byte(err.Message)); writeErr != nil {
		log.Printf("Error writing error message: %v", writeErr)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	r, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("Error parsing request: %v", err)
		WriteHandlerError(conn, &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    "Bad Request\n",
		})
		return
	}

	buf := bytes.NewBuffer([]byte{})

	handlerErr := s.handler(buf, r)
	if handlerErr != nil {
		WriteHandlerError(conn, handlerErr)
		return
	}

	response.WriteStatusLine(conn, response.StatusOK)

	headers := response.GetDefaultHeaders(buf.Len())
	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	s := &Server{listener: l, handler: handler}
	go s.listen()

	return s, nil
}
