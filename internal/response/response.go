package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/TheMaru/http_from_tcp/internal/headers"
)

// type StatusLine struct {
// 	HttpVersion string
// 	StatusCode string
// }
//
// type Response struct {
// 	StatusLine StatusLine
// }

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	httpVersion := "HTTP/1.1"
	var statusString string
	switch statusCode {
	case StatusOK:
		statusString = "OK"
	case StatusBadRequest:
		statusString = "Bad Request"
	case StatusInternalServerError:
		statusString = "Internal Server Error"
	default:
		statusString = ""
	}

	statusLine := fmt.Sprintf("%s %d %s\r\n", httpVersion, statusCode, statusString)

	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers["content-length"] = strconv.Itoa(contentLen)
	headers["connection"] = "close"
	headers["content-type"] = "text/plain"

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	return nil
}
