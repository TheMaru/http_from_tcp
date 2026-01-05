package request

import (
	"errors"
	"io"
	"regexp"
	"strings"
)

type RequestStatus = int

const (
	RequestStateInitialized RequestStatus = iota
	RequestStateDone
)

type Request struct {
	RequestLine RequestLine
	State       RequestStatus
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var upperASCII = regexp.MustCompile(`^[A-Z]+$`)
var bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	request := &Request{
		State: RequestStateInitialized,
	}

	for request.State != RequestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.State = RequestStateDone
				break
			}
			return nil, err
		}

		readToIndex += numBytesRead

		numBytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}

	return request, nil
}

func parseRequestLine(req string) (int, *RequestLine, error) {
	i := strings.Index(req, "\r\n")
	if i == -1 {
		return 0, nil, nil
	}

	line := req[:i]
	bytesConsumed := i + 2

	parts := strings.Split(line, " ")
	if len(parts) == 1 {
		return 0, nil, nil
	}

	if !upperASCII.MatchString(parts[0]) {
		return 0, nil, errors.New("Invalid Method name")
	}

	if !strings.HasPrefix(parts[2], "HTTP/") {
		return 0, nil, errors.New("Not HTTP")
	}
	if !strings.Contains(parts[2], "1.1") {
		return 0, nil, errors.New("Invalid http version")
	}
	versionParts := strings.Split(parts[2], "/")

	requestLine := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   versionParts[1],
	}

	return bytesConsumed, &requestLine, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case RequestStateInitialized:
		n, requestLine, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.State = RequestStateDone
		return n, nil
	case RequestStateDone:

		return 0, errors.New("Error: trying to read data in a done state")
	default:
		return 0, errors.New("Error: unknown state")
	}

}
