package headers

import (
	"errors"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	str := string(data)
	delimiter := "\r\n"

	if !strings.Contains(str, delimiter) {
		return 0, false, nil
	}

	if strings.HasPrefix(string(data), "\r\n") {
		return len(data), true, nil
	}

	dIndex := strings.Index(str, delimiter)

	headerStr := str[:dIndex+len(delimiter)]

	trimmedDataStr := strings.TrimSpace(headerStr)

	firstColonPos := strings.Index(trimmedDataStr, ":")

	if firstColonPos < 1 {
		return 0, false, errors.New("missing field-name")
	}

	fieldName, fieldValue := trimmedDataStr[:firstColonPos], strings.TrimSpace(trimmedDataStr[firstColonPos+1:])

	if strings.HasSuffix(fieldName, " ") {
		return 0, false, errors.New("invalid field-name")
	}

	h[fieldName] = fieldValue

	return len(headerStr), false, nil
}
