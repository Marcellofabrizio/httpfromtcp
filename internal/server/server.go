package server

import (
	"log"
	"net"
)

type Server struct {
	Port     string
	listener net.Listener
}

func Serve(port int) (*Server, error) {

	portStr := ":" + string(port)
	l, err := net.Listen("tcp", portStr)
	if err != nil {
		log.Fatal(err.Error())
	}

	server := Server{
		Port:     portStr,
		listener: l,
	}

	go func() {
		defer server.Close()
		server.listen()
	}()

	return nil, nil

}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) listen() {
	log.Printf("App listening on port %s\n", s.Port)

	conn, err := s.listener.Accept()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) handle(conn net.Conn) {

}
