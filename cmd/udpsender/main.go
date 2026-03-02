package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const port = ":42069"

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost"+port)
	if err != nil {
		log.Fatal(err)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		str, err := r.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		_, err = conn.Write([]byte(str))
		if err != nil {
			log.Println(err)
			return
		}
	}
}
