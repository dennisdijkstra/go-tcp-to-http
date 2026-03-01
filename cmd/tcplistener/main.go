package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Listening for TCP traffic on", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Printf("New connection from %s\n", conn.RemoteAddr())

		lines := getLinesChannel(conn)

		for line := range lines {
			fmt.Println(line)
		}

		fmt.Printf("Connection from %s closed\n", conn.RemoteAddr())
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		defer f.Close()

		buff := make([]byte, 8)
		currLine := ""

		for {
			n, err := f.Read(buff)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Println(err)
				break
			}

			str := string(buff[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- currLine + parts[i]
				currLine = ""
			}
			currLine += parts[len(parts)-1]
		}
		if currLine != "" {
			lines <- currLine
		}
	}()
	return lines
}
