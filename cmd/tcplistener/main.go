package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		fmt.Println("connection has been accepted")

		ch := getLinesChannel(conn)

		for line := range ch {
			fmt.Println(line)
		}
		fmt.Println("connection is closed again")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer f.Close()
		defer close(ch)
		buffer := make([]byte, 8)
		line := ""
		for {
			n, err := f.Read(buffer)
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Printf("read error: %v\n", err)
				return
			}
			if n > 0 {
				parts := strings.Split(string(buffer[:n]), "\n")
				for i := 0; i < len(parts)-1; i++ {
					ch <- line + parts[i]
					line = ""
				}
				line += parts[len(parts)-1]
			}
		}
	}()

	return ch
}
