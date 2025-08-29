package response

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
)

const (
	OK                    = 200
	BAD_REQUEST           = 400
	INTERNAL_SERVER_ERROR = 500
)

type Writer struct {
	StatusLine []byte
	Headers    []byte
	Body       []byte
}

type StatusCode int

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var statusLine []byte
	switch statusCode {
	case OK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
		break
	case BAD_REQUEST:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
		break
	case INTERNAL_SERVER_ERROR:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
		break
	default:
		statusLine = []byte("\r\n")
	}

	w.StatusLine = statusLine

	return nil
}

func GetDefaultHeaders(contentLength int) headers.Headers {
	defaultHeaders := headers.Headers{}

	defaultHeaders["Content-Length"] = fmt.Sprintf("%d", contentLength)
	defaultHeaders["Connection"] = "close"
	defaultHeaders["Content-Type"] = "text/html"

	return defaultHeaders
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	headerByteBuffer := bytes.NewBuffer([]byte{})
	for k, v := range headers {
		headerByteBuffer.Write([]byte(fmt.Sprintf("%s:%s\r\n", k, v)))
	}
	headerByteBuffer.Write([]byte("\r\n"))

	w.Headers = headerByteBuffer.Bytes()

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	httpVersion := []byte("HTTP/1.1 ")
	statusLine := w.StatusLine
	titleIndex := bytes.Index(w.StatusLine, httpVersion)
	if titleIndex == -1 {
		return 0, fmt.Errorf("error: status line is malformed")
	}

	title := statusLine[titleIndex+len(httpVersion):]
	h1 := statusLine[titleIndex+len(httpVersion)+4:]

	w.Body = []byte(fmt.Sprintf("<html>\n "+
		" <head>\n    "+
		"<title>%s</title>\n "+
		" </head>\n "+
		" <body>\n   "+
		" <h1>%s</h1>\n   "+
		" <p>%s</p>\n "+
		" </body>\n</html>", string(title), string(h1), string(p)))

	return len(w.Body), nil
}
