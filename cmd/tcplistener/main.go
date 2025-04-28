package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {

	out := make(chan string)

	go func() {
		defer close(out)
		buffer := make([]byte, 8, 8)
		currentLine := ""

		for {
			bytesRead, eofErr := f.Read(buffer)

			if eofErr != nil {

				if currentLine != "" {
					out <- currentLine
				}

				return
			}

			str := string(buffer[:bytesRead])
			parts := strings.Split(str, "\n")

			for i := 0; i < len(parts)-1; i++ {
				out <- fmt.Sprintf("%s%s", currentLine, parts[i])
				currentLine = ""
			}

			currentLine += parts[len(parts)-1]
		}
	}()

	return out
}

func main() {

	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()

	if err != nil {
		log.Fatal(err)
	}

	for {
		log.Println("Waiting connection...")
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Connection accepted")

		request, err := request.RequestFromReader(conn)

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)
		
		fmt.Println("Headers:")

		for k, v := range request.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}

	}
}
