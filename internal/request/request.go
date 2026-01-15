package request

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/TheMaru/http_from_tcp/internal/headers"
)

type RequestStatus = int

const (
	requestStateInitialized RequestStatus = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       RequestStatus
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
		Headers: headers.NewHeaders(),
		state:   requestStateInitialized,
	}

	for request.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", request.state, numBytesRead)
				}
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
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		n, requestLine, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, noMoreHeaders, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		if noMoreHeaders {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		expectedContentLength, exists := r.Headers.Get("Content-Length")
		if !exists {
			r.state = requestStateDone
			return 0, nil
		}

		expectedContentLengthInt, err := strconv.Atoi(expectedContentLength)
		if err != nil {
			return 0, err
		}

		r.Body = append(r.Body, data...)

		if len(r.Body) > expectedContentLengthInt {
			return len(data), fmt.Errorf("Body exceeds expected length: expected %d, got %d", expectedContentLengthInt, len(r.Body))
		}
		if len(r.Body) == expectedContentLengthInt {
			r.state = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, errors.New("Error: trying to read data in a done state")
	default:
		return 0, errors.New("Error: unknown state")
	}
}
