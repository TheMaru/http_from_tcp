package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal("Can't read file", err)
	}

	ch := getLinesChannel(file)
	for line := range ch {
		fmt.Printf("read: %v\n", line)
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
