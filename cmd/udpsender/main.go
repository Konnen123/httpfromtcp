package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatalf("error to resolve udp addr")
		return
	}

	udp, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("error dial udp")
		return
	}

	stdinReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")

		readString, err := stdinReader.ReadString('\n')
		if err != nil {
			log.Fatalf("error: %s", err.Error())
		}

		_, err = udp.Write([]byte(readString))
		if err != nil {
			return
		}
	}
}
