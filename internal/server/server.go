package server

import (
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/TheMaru/http_from_tcp/internal/response"
)

type Server struct {
	Listener     net.Listener
	ServerClosed atomic.Bool
}

func Serve(port int) (*Server, error) {
	portString := strconv.Itoa(port)
	listener, err := net.Listen("tcp", ":"+portString)
	if err != nil {
		return nil, fmt.Errorf("Listener on network with port %d could not be created: %v", port, err)
	}

	server := Server{
		Listener: listener,
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
	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		fmt.Printf("Error during writing of response status line: %v\n", err)
	}
	headers := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		fmt.Printf("Error during writing of response headers: %v\n", err)
	}
	fmt.Fprintf(conn, "\r\n")
	fmt.Fprintf(conn, "Hello World!\n")
}
