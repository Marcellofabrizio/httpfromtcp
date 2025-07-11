package server

import (
	"bytes"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type HandlerError struct {
	Message    string
	StatusCode response.StatusCode
}

func (hErr *HandlerError) Write(w response.Writer) error {
	err := w.WriteStatusLine(hErr.StatusCode)

	if err != nil {
		return err
	}

	body := []byte(hErr.Message)

	headers := response.GetDefaultHeaders(len(body), "text/html")
	if err := w.WriteHeaders(headers); err != nil {
		return err
	}

	_, err = w.WriteBody(body)
	return err
}

type Handler func(w *response.Writer, req *request.Request) *HandlerError

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

	buf := bytes.NewBuffer([]byte{})
	writer := &response.Writer{
		Buffer: buf,
	}

	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(*writer)
		conn.Write(writer.Buffer.Bytes())
		return
	}

	hErr := s.handler(writer, req)

	if hErr != nil {
		hErr.Write(*writer)
		conn.Write(writer.Buffer.Bytes())
		return
	}

	conn.Write(writer.Buffer.Bytes())
}
