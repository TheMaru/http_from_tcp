package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/TheMaru/http_from_tcp/internal/request"
	"github.com/TheMaru/http_from_tcp/internal/response"
)

type Server struct {
	Listener     net.Listener
	ServerClosed atomic.Bool
	handler      Handler
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func Serve(port int, handler Handler) (*Server, error) {
	portString := strconv.Itoa(port)
	listener, err := net.Listen("tcp", ":"+portString)
	if err != nil {
		return nil, fmt.Errorf("Listener on network with port %d could not be created: %v", port, err)
	}

	server := Server{
		Listener: listener,
		handler:  handler,
	}
	server.ServerClosed.Store(false)

	go server.listen()

	return &server, nil
}

func (s *Server) Close() error {
	s.ServerClosed.Store(true)
	err := s.Listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.ServerClosed.Load() {
				return
			}
			fmt.Printf("New connection could not be accepted: %v\n", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}

	b := buf.Bytes()
	headers := response.GetDefaultHeaders(len(b))

	response.WriteStatusLine(conn, response.StatusOK)
	response.WriteHeaders(conn, headers)
	conn.Write(b)
}
