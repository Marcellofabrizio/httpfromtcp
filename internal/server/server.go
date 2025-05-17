package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	Port     string
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {

	log.Println(port)
	portStr := ":" + strconv.Itoa(port)
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
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	log.Printf("App listening on port %s\n", s.Port)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				log.Println("Server closed, stopping accept loop")
				return
			}
			log.Printf("Accept error: %v\n", err)
			continue
		}

		log.Println("Connection Accepted")
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	_, err := request.RequestFromReader(conn)

	if err != nil {
		log.Printf("Failed to parse request: %v\n", err)
		conn.Close()
		return
	}

	err = response.WriteStatusLine(conn, 200)

	if err != nil {
		log.Printf("failed to write status line: %v\n", err)
	}

	headers := response.GetDefaultHeaders(0)

	err = response.WriteHeaders(conn, headers)

	for k, v := range headers {
		fmt.Printf("%s: %s\n", k, v)
	}

	if err != nil {
		log.Printf("failed to write response: %v\n", err)
	}
}
