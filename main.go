package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	f, err := os.Open("./messages.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	buf := make([]byte, 8)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("read: %s\n", string(buf[:n]))
	}
}
