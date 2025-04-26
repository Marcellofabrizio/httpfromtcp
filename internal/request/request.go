package request

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strings"
	"unicode"
)

type RequestStatus int

const (
	Initialized RequestStatus = iota
	Done
	ParsingHeaders
)

type Request struct {
	RequestLine RequestLine
	Status      RequestStatus
	Headers     headers.Headers
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {

	totalBytesConsumed := 0

	if r.Status == Initialized {
		parsedLine, bytesConsumed, err := parseRequestLine(string(data))

		if err != nil {
			return -1, err
		} else if bytesConsumed == 0 {
			return 0, nil
		}

		r.RequestLine = *parsedLine
		r.Status = ParsingHeaders
		totalBytesConsumed = bytesConsumed
		return bytesConsumed, nil
	}

	for r.Status != Done {
		n, done, err := r.Headers.Parse(data[totalBytesConsumed:])

		if err != nil {
			return 0, err
		}

		if n == 0 {
			return 0, nil
		}

		totalBytesConsumed += n

		if !done {
			return totalBytesConsumed, nil
		} else {
			r.Status = Done
			return totalBytesConsumed, nil
		}
	}

	if r.Status == Done {
		return -1, errors.New("cannot read on complete state")
	}

	return -1, errors.New("unknown error")
}

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {

	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0

	request := Request{
		Status:  Initialized,
		Headers: headers.NewHeaders(),
	}

	for request.Status != Done {
		bytesRead, err := reader.Read(buf[readToIndex:])

		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		if err != nil && err != io.EOF {
			return nil, err
		}

		readToIndex += bytesRead

		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		parsed, perr := request.parse(buf[:readToIndex])
		if perr != nil {
			return nil, perr
		}

		if parsed > 0 {
			copy(buf, buf[parsed:readToIndex])
			readToIndex -= parsed
		}

		if err == io.EOF {
			break
		}

	}

	return &request, nil
}

func parseRequestLine(content string) (*RequestLine, int, error) {

	delimiter := "\r\n"
	if !strings.Contains(content, delimiter) {
		return nil, 0, nil
	}

	dIndex := strings.Index(content, delimiter)

	requestLineContent := content[:dIndex+len(delimiter)]

	parts := strings.Split(requestLineContent, " ")

	if len(parts) != 3 {
		return nil, 0, errors.New("invalid request line size")
	}

	method, target, httpVersion := parts[0], parts[1], parts[2]

	isMethodValid := validateMethod(method)

	if !isMethodValid {
		return nil, 0, errors.New(fmt.Sprintf("invalid method %s", method))
	}

	isTargetValid := validateTarget(target)

	if !isTargetValid {
		return nil, 0, errors.New(fmt.Sprintf("invalid target %s", target))
	}

	versionNumber := strings.TrimSpace(strings.Split(httpVersion, "/")[1])

	isValidVersion := validateVersion(versionNumber)

	if !isValidVersion {
		return nil, 0, errors.New(fmt.Sprintf("invalid version %s", httpVersion))
	}

	requestLine := RequestLine{
		HttpVersion:   versionNumber,
		RequestTarget: target,
		Method:        method,
	}

	return &requestLine, len(requestLineContent), nil
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
