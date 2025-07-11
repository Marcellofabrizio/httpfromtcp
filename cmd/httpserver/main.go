package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {

	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) *server.HandlerError {

		log.Println(req.RequestLine.RequestTarget)
		if req.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				Message: `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`,
				StatusCode: response.StatusBadRequest,
			}
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				Message:    "My problem is not your problem\n",
				StatusCode: response.StatusInternalServerError,
			}
		} else {
			w.WriteStatusLine(response.StatusOK)
			body := []byte("All good, frfr")
			headers := response.GetDefaultHeaders(len(body), "text/html")
			w.WriteHeaders(headers)
			w.WriteBody(body)
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
