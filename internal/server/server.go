package server

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"net"
	"sync/atomic"
)

type Server struct {
	IsTerminated atomic.Bool
	Listener     net.Listener
	Handler      Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    []byte
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	var isTerminated atomic.Bool
	isTerminated.Store(false)

	server := &Server{
		IsTerminated: isTerminated,
		Listener:     listener,
		Handler:      handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	if s.IsTerminated.Load() {
		return fmt.Errorf("error: server is already terminated")
	}

	s.IsTerminated.Store(true)
	err := s.Listener.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) listen() {
	for {
		accept, err := s.Listener.Accept()
		if err != nil {
			return
		}

		go s.handle(accept)
	}
}

func (s *Server) handle(conn net.Conn) {
	var statusCode response.StatusCode
	statusCode = response.OK
	req, err := request.RequestFromReader(conn)

	defer conn.Close()
	if err != nil {
		fmt.Printf("error:%s", err.Error())

		herr := &HandlerError{
			StatusCode: response.BAD_REQUEST,
			Message:    []byte(err.Error()),
		}
		herr.Write(conn)
		return
	}

	var writer response.Writer

	buffer := bytes.NewBuffer([]byte{})
	handlerError := s.Handler(buffer, req)
	if handlerError != nil {
		handlerError.Write(conn)
		return
	}
	body := buffer.Bytes()
	writer.WriteStatusLine(statusCode)
	bodyLength, err := writer.WriteBody(body)
	if err != nil {
		return
	}

	defaultHeaders := response.GetDefaultHeaders(bodyLength)
	writer.WriteHeaders(defaultHeaders)

	conn.Write(writer.StatusLine)
	conn.Write(writer.Headers)
	conn.Write(writer.Body)
}

func (h *HandlerError) Write(conn io.Writer) {
	var writer response.Writer
	statusCode := h.StatusCode
	message := h.Message
	writer.WriteStatusLine(statusCode)
	writer.WriteBody(message)
	defaultHeaders := response.GetDefaultHeaders(len(writer.Body))
	writer.WriteHeaders(defaultHeaders)

	conn.Write(writer.StatusLine)
	conn.Write(writer.Headers)
	conn.Write(writer.Body)
}
