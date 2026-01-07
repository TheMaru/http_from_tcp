package headers

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

var ErrInvalidHeader = errors.New("Header formatting error")
var headerKeyRE = regexp.MustCompile(`^[A-Za-z0-9!#$%&'\*\+\-\.\^_` + "`" + `\|~]+$`)

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
	isValidKey := headerKeyRE.MatchString(key)
	if !isValidKey {
		n = 0
		err = errors.New("invalid characters in header key")
		return
	}

	startValue := colonIndex + 1
	value := strings.Trim(header[startValue:], " ")
	h[strings.ToLower(key)] = value

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
