package headers

import (
	"errors"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	if !strings.HasSuffix(string(data), "\r\n") {
		return 0, false, nil
	}

	if strings.HasPrefix(string(data), "\r\n") {
		return len(data), true, nil
	}

	trimmedDataStr := strings.TrimSpace(string(data))

	firstColonPos := strings.Index(trimmedDataStr, ":")

	if firstColonPos < 1 {
		return 0, false, errors.New("missing field-name")
	}

	fieldName, fieldValue := trimmedDataStr[:firstColonPos], strings.TrimSpace(trimmedDataStr[firstColonPos+1:])

	fmt.Printf("field-name: %s\nfield-value: %s\n", fieldName, fieldValue)

	return 0, true, nil
}
