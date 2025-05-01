package request

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type RequestStatus int

const (
	Initialized RequestStatus = iota
	ParsingHeaders
	ParsingBody
	Done
)

type Request struct {
	RequestLine RequestLine
	Status      RequestStatus
	Headers     headers.Headers
	Body        []byte
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
		switch r.Status {
		case ParsingHeaders:
			n, done, err := r.Headers.Parse(data[totalBytesConsumed:])

			if err != nil {
				return 0, err
			}

			if n == 0 {
				return 0, nil
			}

			if done {
				r.Status = ParsingBody
			}

			totalBytesConsumed += n
			return totalBytesConsumed, nil
		case ParsingBody:
			remaining := data[totalBytesConsumed:]
			value, ok := r.Headers.Get("Content-Length")
			if !ok {
				r.Status = Done
				return totalBytesConsumed, nil
			}

			contentLength, err := strconv.Atoi(value)
			if err != nil {
				return 0, errors.New("invalid content-length")
			}

			if len(remaining) == 0 {
				return 0, nil
			}

			r.Body = append(r.Body, remaining...)
			totalBytesConsumed += len(remaining)

			if len(r.Body) > contentLength {
				return 0, errors.New(fmt.Sprintf("body exceeds content-length. Expected size %d, got %d", contentLength, len(r.Body)))
			}

			if len(r.Body) == contentLength {
				r.Status = Done
			}

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
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := Request{
		Status:  Initialized,
		Headers: headers.NewHeaders(),
	}

	for request.Status != Done {
		if readToIndex > 0 {
			parsed, perr := request.parse(buf[:readToIndex])
			if perr != nil {
				return nil, perr
			}

			if parsed > 0 {
				copy(buf, buf[parsed:readToIndex])
				readToIndex -= parsed
				continue
			}
		}

		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		bytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil && err != io.EOF {
			return nil, err
		}

		if bytesRead == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		readToIndex += bytesRead
	}

	if request.Body != nil && request.Status != Done {
		return nil, errors.New("incomplete body")
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
