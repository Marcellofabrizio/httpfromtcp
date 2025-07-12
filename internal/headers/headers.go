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

func (h Headers) Get(key string) (value string, ok bool) {

	lowerStr := strings.ToLower(key)

	existingValue, ok := h[lowerStr]

	return existingValue, ok
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	str := string(data)
	delimiter := "\r\n"

	if !strings.Contains(str, delimiter) {
		return 0, false, nil
	}

	if strings.HasPrefix(string(data), delimiter) {
		return len(delimiter), true, nil
	}

	dIndex := strings.Index(str, delimiter)

	headerStr := str[:dIndex+len(delimiter)]

	trimmedDataStr := strings.TrimSpace(headerStr)

	firstColonPos := strings.Index(trimmedDataStr, ":")

	if firstColonPos < 1 {
		return 0, false, errors.New("malformed header")
	}

	fieldName, fieldValue := trimmedDataStr[:firstColonPos], strings.TrimSpace(trimmedDataStr[firstColonPos+1:])

	if !validateKey(fieldName) {
		return 0, false, errors.New("invalid field-name")
	}

	fieldName = strings.ToLower(fieldName)

	existingFieldValue, ok := h[fieldName]

	if ok {
		fieldValue = existingFieldValue + ", " + fieldValue
	}

	h[fieldName] = fieldValue

	return len(headerStr), false, nil
}

func (h Headers) Override(key string, value string) {

	_, exists := h.Get(key)

	if exists {
		h.Parse([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
	}

}

func validateKey(key string) bool {

	if strings.HasSuffix(key, " ") {
		return false
	}

	for _, char := range key {

		switch {
		case char >= 'a' && char <= 'z':
			continue
		case char >= 'A' && char <= 'Z':
			continue
		case char >= '0' && char <= '9':
			continue
		case isAllowedSpecialChar(char):
			continue
		default:
			return false
		}
	}

	return true
}

func isAllowedSpecialChar(ch rune) bool {
	switch ch {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return false
	}
}
