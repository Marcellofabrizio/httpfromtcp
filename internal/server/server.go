package server

import (
	"bytes"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type HandlerError struct {
	Message    string
	StatusCode response.StatusCode
}

func (hErr *HandlerError) Write(w io.Writer) error {
	body := []byte(hErr.Message)
	err := response.WriteStatusLine(w, hErr.StatusCode)

	if err != nil {
		return err
	}

	headers := response.GetDefaultHeaders(len(body))
	if err := response.WriteHeaders(w, headers); err != nil {
		return err
	}

	_, err = w.Write(body)
	return err
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	Port     string
	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {

	portStr := ":" + strconv.Itoa(port)
	l, err := net.Listen("tcp", portStr)
	if err != nil {
		log.Fatal(err.Error())
	}

	server := &Server{
		Port:     portStr,
		listener: l,
		handler:  handler,
	}

	go server.listen()
	return server, nil

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
	defer conn.Close()
	req, err := request.RequestFromReader(conn)

	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, req)

	if hErr != nil {
		hErr.Write(conn)
		return
	}

	b := buf.Bytes()
	response.WriteStatusLine(conn, response.StatusOK)
	headers := response.GetDefaultHeaders(len(b))

	err = response.WriteHeaders(conn, headers)

	if err != nil {
		log.Printf("failed to write response: %v\n", err)
	}

	conn.Write(b)
}
