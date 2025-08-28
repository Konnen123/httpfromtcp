package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"net"
)

func main() {

	listen, err := net.Listen("tcp", ":42069")
	if err != nil {
		return
	}

	defer listen.Close()

	for {
		accept, err := listen.Accept()
		if err != nil {
			return
		}

		fmt.Println("A connection has been accepted!")

		parsedRequest, err := request.RequestFromReader(accept)
		if err != nil {
			return
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", parsedRequest.RequestLine.Method)
		fmt.Printf("- Target: %s\n", parsedRequest.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", parsedRequest.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for k, v := range parsedRequest.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}

	}

}
