package request

import (
	"errors"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var upperASCII = regexp.MustCompile(`^[A-Z]+$`)

func RequestFromReader(reader io.Reader) (*Request, error) {
	req, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	reqString := string(req)
	lines := strings.Split(reqString, "\r\n")
	reqLine, err := parseRequestLine(lines[0])
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *reqLine,
	}, nil
}

func parseRequestLine(line string) (*RequestLine, error) {
	parts := strings.Split(line, " ")

	if !upperASCII.MatchString(parts[0]) {
		return nil, errors.New("Invalid Method name")
	}

	if !strings.HasPrefix(parts[2], "HTTP/") {
		return nil, errors.New("Not HTTP")
	}
	if !strings.Contains(parts[2], "1.1") {
		return nil, errors.New("Invalid http version")
	}
	versionParts := strings.Split(parts[2], "/")

	requestLine := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   versionParts[1],
	}

	return &requestLine, nil
}
