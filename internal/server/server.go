package server

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"net"
	"strings"
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

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		s.handleChunkedResponse(conn, req)
	} else {
		s.handleNormalResponse(conn, req)
	}

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

func (s *Server) handleNormalResponse(conn net.Conn, req *request.Request) {
	var writer response.Writer

	buffer := bytes.NewBuffer([]byte{})

	handlerError := s.Handler(buffer, req)
	if handlerError != nil {
		handlerError.Write(conn)
		return
	}
	body := buffer.Bytes()

	writer.WriteStatusLine(response.OK)

	var responseHeaders headers.Headers
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/video") {
		writer.Body = body
		responseHeaders = response.GetVideoHeaders(len(body))
	} else {
		bodyLength, err := writer.WriteBody(body)
		if err != nil {
			return
		}
		responseHeaders = response.GetDefaultHeaders(bodyLength)
	}
	writer.WriteHeaders(responseHeaders)

	conn.Write(writer.StatusLine)
	conn.Write(writer.Headers)
	conn.Write(writer.Body)
}

func (s *Server) handleChunkedResponse(conn net.Conn, req *request.Request) {
	var writer response.Writer

	writer.WriteStatusLine(response.OK)
	chunkedHeaders := response.GetChunkedHeaders()
	writer.WriteHeaders(chunkedHeaders)

	buffer := bytes.NewBuffer([]byte{})
	go s.Handler(buffer, req)

	for {
		data := make([]byte, 1024)
		readBytes, err := buffer.Read(data)
		if err != nil {
			continue
		}
		dataRead := data[:readBytes]
		if bytes.Equal(dataRead, []byte("\r\n")) {
			writer.WriteChunkedBodyDone()
			break
		}
		writer.WriteChunkedBody(data[:readBytes])
	}
	conn.Write(writer.Body)

	trailers := headers.Headers{}
	trailers["X-Content-SHA256"] = fmt.Sprintf("%v", sha256.Sum256(writer.Body))
	trailers["X-Content-Length"] = fmt.Sprintf("%v", len(writer.Body))

	writer.WriteTrailers(trailers)
	conn.Write(writer.Trailers)
}
