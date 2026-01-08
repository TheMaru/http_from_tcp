package main

import (
	"fmt"
	"log"
	"net"

	"github.com/TheMaru/http_from_tcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("Can't create listener", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Can't accept connection", err)
		}
		fmt.Println("Connection has been accepted")

		request, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println("Error while reading data", err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range request.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("Connection is closed again")
	}
}
