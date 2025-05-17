package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	httpVersion := "HTTP/1.1"
	switch statusCode {
	case StatusOK:
		_, err := w.Write([]byte(fmt.Sprintf("%s %d %s\r\n", httpVersion, StatusOK, "OK")))
		return err
	case StatusBadRequest:
		_, err := w.Write([]byte(fmt.Sprintf("%s %d %s\r\n", httpVersion, StatusBadRequest, "Bad Request")))
		return err
	case StatusInternalServerError:
		_, err := w.Write([]byte(fmt.Sprintf("%s %d %s\r\n", httpVersion, StatusInternalServerError, "Internal Server Error")))
		return err
	default:
		_, err := w.Write([]byte(fmt.Sprintf("%s %d\r\n", httpVersion, statusCode)))
		return err
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers.Parse([]byte(fmt.Sprintf("Content-Length: %d\r\n", contentLen)))
	headers.Parse([]byte("Connection: close\r\n"))
	headers.Parse([]byte("Content-Type: text/plain\r\n"))

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {

	for k, v := range headers {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))

	if err != nil {
		return err
	}

	return nil
}
