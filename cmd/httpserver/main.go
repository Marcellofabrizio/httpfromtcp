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
				Message: `
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`,
				ContentType: "text/html",
				StatusCode:  response.StatusBadRequest,
			}
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				Message: `
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
				`,
				ContentType: "text/html",
				StatusCode:  response.StatusInternalServerError,
			}
		} else {
			w.WriteStatusLine(response.StatusOK)
			body := []byte(`
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
			headers := response.GetDefaultHeaders(len(body))
			headers.Override("Content-Type", "text/html")
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
