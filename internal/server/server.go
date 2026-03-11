package server

import (
	"log"
	"net"
	"strconv"

	"github.com/dennisdijkstra/go-tcp-to-http/internal/response"
)

type Server struct {
	listener net.Listener
}

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

	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Printf("error writing status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(0)

	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Printf("error writing header: %v", err)
		return
	}
}

func Serve(port int) (*Server, error) {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	s := &Server{listener: l}
	go s.listen()

	return s, nil
}
