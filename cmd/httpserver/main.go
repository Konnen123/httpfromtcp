package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	srv, err := server.Serve(port, Handler)
	if err != nil {
		log.Fatalf("Error starting srv: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func Handler(w io.Writer, req *request.Request) *server.HandlerError {
	requestPath := req.RequestLine.RequestTarget

	switch requestPath {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: 400,
			Message:    []byte("Your problem is not my problem"),
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: 500,
			Message:    []byte("Woopsie, my bad"),
		}
	case "/video":
		file, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			fmt.Printf("Error: %s", err.Error())
			return &server.HandlerError{
				StatusCode: 500,
				Message:    []byte("Woopsie, my bad"),
			}
		}

		w.Write(file)
	default:
		substr := "/httpbin"
		httpBinPrefixIndex := strings.Index(requestPath, substr)
		if httpBinPrefixIndex != -1 {
			trimPath := requestPath[httpBinPrefixIndex+len(substr):]

			getResponse, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", trimPath))
			if err != nil {
				return &server.HandlerError{
					StatusCode: 500,
					Message:    []byte("Error at getting httpbin"),
				}
			}

			bufferSize := 1024
			for {
				dataBuffer := make([]byte, bufferSize)
				n, err := getResponse.Body.Read(dataBuffer)
				if err != nil {
					break
				}

				fmt.Printf("Read size: %d\n", n)

				w.Write(dataBuffer[:bufferSize])
			}
			w.Write([]byte("\r\n"))
			break
		}

		w.Write([]byte("All good, frfr"))
	}

	return nil
}
