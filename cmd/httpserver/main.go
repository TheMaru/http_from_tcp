package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/TheMaru/http_from_tcp/internal/headers"
	"github.com/TheMaru/http_from_tcp/internal/request"
	"github.com/TheMaru/http_from_tcp/internal/response"
	"github.com/TheMaru/http_from_tcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w *response.Writer, req *request.Request) {
		var status response.StatusCode
		var body string

		if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			destPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
			resp, err := http.Get("https://httpbin.org/" + destPath)
			if err != nil {
				log.Fatalf("Error proxying: %v", err)
			}
			defer resp.Body.Close()

			h := response.GetDefaultHeaders(0)
			h.Delete("content-length")
			h.Set("transfer-encoding", "chunked")
			h.Set("content-type", resp.Header.Get("Content-Type"))
			h.Set("Trailer", "X-Content-SHA256, X-Content-Length")

			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(h)

			var fullBody []byte
			buf := make([]byte, 1024)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					fullBody = append(fullBody, buf[:n]...)
					w.WriteChunkedBody(buf[:n])
				}
				if err != nil {
					break
				}
			}
			w.WriteChunkedBodyDone()
			hash := sha256.Sum256(fullBody)
			trailers := headers.NewHeaders()
			trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
			trailers.Set("X-Content-Length", strconv.Itoa(len(fullBody)))
			w.WriteTrailers(trailers)

			return
		}

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			status = response.StatusBadRequest
			body = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
		case "/myproblem":
			status = response.StatusInternalServerError
			body = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
		default:
			status = response.StatusOK
			body = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
		}

		b := []byte(body)
		h := response.GetDefaultHeaders(len(b))
		h.Set("content-type", "text/html")

		w.WriteStatusLine(status)
		w.WriteHeaders(h)
		w.WriteBody(b)
	}

	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
