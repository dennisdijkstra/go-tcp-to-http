package server

import (
	"fmt"
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

type Handler func(w *response.Writer, req *request.Request)

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

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	w := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		body := []byte(fmt.Sprintf("Error parsing request: %v", err))
		if err := w.WriteStatusLine(response.StatusBadRequest); err != nil {
			log.Printf("Error writing status line: %v", err)
			return
		}

		headers := response.GetDefaultHeaders(len(body))
		if err := w.WriteHeaders(headers); err != nil {
			log.Printf("Error writing headers: %v", err)
			return
		}

		if _, err := w.WriteBody(body); err != nil {
			log.Printf("Error writing response body: %v", err)
			return
		}
		return
	}

	s.handler(w, r)
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
