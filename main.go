package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		defer f.Close()

		buffer := make([]byte, 8)
		currLine := ""

		for {
			n, err := f.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Println(err)
				break
			}

			str := string(buffer[:n])
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

func main() {
	f, err := os.Open("./messages.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	lines := getLinesChannel(f)

	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}
