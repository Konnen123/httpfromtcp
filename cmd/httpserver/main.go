package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
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
	default:
		w.Write([]byte("All good, frfr"))
	}

	return nil
}
