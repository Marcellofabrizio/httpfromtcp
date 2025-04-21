package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	msgChannel := getLinesChannel(reader)

	requestLineContent := <-msgChannel

	requestLine, err := parseRequestLine(requestLineContent)

	if err != nil {
		return nil, err
	}

	request := Request{
		RequestLine: *requestLine,
	}

	return &request, nil
}

func parseRequestLine(content string) (*RequestLine, error) {

	parts := strings.Split(content, " ")

	if len(parts) != 3 {
		return nil, errors.New("invalid request line size")
	}

	method, target, httpVersion := parts[0], parts[1], parts[2]

	isMethodValid := validateMethod(method)

	if !isMethodValid {
		return nil, errors.New(fmt.Sprintf("invalid method %s", method))
	}

	isTargetValid := validateTarget(target)

	if !isTargetValid {
		return nil, errors.New(fmt.Sprintf("invalid target %s", target))
	}

	versionNumber := strings.Split(httpVersion, "/")[1]

	isValidVersion := validateVersion(versionNumber)

	if !isValidVersion {
		return nil, errors.New(fmt.Sprintf("invalid version %s", httpVersion))
	}

	requestLine := RequestLine{
		HttpVersion:   versionNumber,
		RequestTarget: target,
		Method:        method,
	}

	return &requestLine, nil
}

func validateMethod(s string) bool {

	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}

	return true
}

func validateTarget(s string) bool {

	return strings.HasPrefix(s, "/")
}

func validateVersion(s string) bool {

	return s == "1.1"
}

func getLinesChannel(f io.Reader) <-chan string {

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
			parts := strings.Split(str, "\r\n")

			for i := 0; i < len(parts)-1; i++ {
				out <- fmt.Sprintf("%s%s", currentLine, parts[i])
				currentLine = ""
			}

			currentLine += parts[len(parts)-1]
		}
	}()

	return out
}
