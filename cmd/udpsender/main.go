package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	address, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("Address not available: %v", err)
	}

	udpConn, err := net.DialUDP("udp", nil, address)
	if err != nil {
		log.Fatalf("Connection not possible %v", err)
	}
	defer udpConn.Close()

	buffReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		line, err := buffReader.ReadString('\n')
		if err != nil {
			log.Fatalf("No delimiter found %v", err)
		}

		_, err = udpConn.Write([]byte(line))
		if err != nil {
			fmt.Printf("Error during sending %v\n", err)
		}
	}
}
