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

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
)

func (s writerState) String() string {
	switch s {
	case writerStateStatusLine:
		return "writerStateStatusLine"
	case writerStateHeaders:
		return "writerStateHeaders"
	case writerStateBody:
		return "writerStateBody"
	default:
		return fmt.Sprintf("writerState(%d)", int(s))
	}
}

type Writer struct {
	dest  io.Writer
	state writerState
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateStatusLine {
		return fmt.Errorf("cannot write status line in state %s", w.state)
	}

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

	_, err := w.dest.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	w.state = writerStateHeaders
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers["content-length"] = strconv.Itoa(contentLen)
	headers["connection"] = "close"
	headers["content-type"] = "text/plain"

	return headers
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writerStateHeaders {
		return fmt.Errorf("cannot write headers in state %s", w.state)
	}

	for key, value := range headers {
		_, err := fmt.Fprintf(w.dest, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	fmt.Fprint(w.dest, "\r\n")

	w.state = writerStateBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %s", w.state)
	}
	return w.dest.Write(p)
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		dest: w,
	}
}
