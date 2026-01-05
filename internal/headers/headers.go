package headers

import (
	"errors"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

var ErrInvalidHeader = errors.New("Header formatting error")

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	dataString := string(data)
	done = false
	err = nil
	n = 0

	i := strings.Index(dataString, "\r\n")
	if i == -1 {
		return
	}
	// new line at the start of a line means end of headers found
	if i == 0 {
		n = 2
		done = true
		return
	}

	header := dataString[:i]

	startKey := skipLeadingWhitespace(header)
	colonIndex := strings.Index(header, ":")
	key := header[startKey:colonIndex]
	if containsWhitespace(key) {
		n = 0
		err = ErrInvalidHeader
		return
	}

	startValue := colonIndex + 1
	value := strings.Trim(header[startValue:], " ")
	h[key] = value

	n = len(dataString[:i]) + 2
	return
}

func skipLeadingWhitespace(s string) int {
	i := 0
	for i < len(s) {
		r := rune(s[i])
		if !unicode.IsSpace(r) {
			break
		}
		i++
	}
	return i
}

func containsWhitespace(s string) bool {
	containsSpace := false
	i := 0
	for i < len(s) {
		r := rune(s[i])
		if unicode.IsSpace(r) {
			containsSpace = true
			break
		}
		i++
	}
	return containsSpace
}
