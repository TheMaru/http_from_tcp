package server

import (
	"fmt"
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

type Handler func(w *response.Writer, req *request.Request)

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
	w := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		body := []byte(err.Error())
		w.WriteStatusLine(response.StatusBadRequest)
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return
	}

	s.handler(w, req)
}
