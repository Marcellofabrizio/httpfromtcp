package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {

	server, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {

		log.Println(req.RequestLine.RequestTarget)
		if req.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				Message:    "Your problem is not my problem\n",
				StatusCode: response.StatusBadRequest,
			}
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				Message:    "My problem is not your problem\n",
				StatusCode: response.StatusInternalServerError,
			}
		} else {
			w.Write([]byte("All good, frfr\n"))
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
