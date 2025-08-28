package request

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var (
	requestData []byte
	bytesRead   int
	bytesParsed int
)

const (
	REQUEST_STATE_DONE            = 2
	REQUEST_STATE_PARSING_HEADERS = 1
	REQUEST_STATE_INITIALIZED     = 0
	SEPARATOR                     = "\r\n"
)

func RequestFromReader(reader io.Reader) (*Request, error) {

	bytesRead = 0
	bytesParsed = 0
	requestData = make([]byte, 8)
	request := &Request{
		State:   REQUEST_STATE_INITIALIZED,
		Headers: headers.Headers{},
	}

	isEOF := false

	for {
		data := make([]byte, 8)
		n, err := reader.Read(data)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			isEOF = true
		}

		bytesRead += n

		parse, err := request.parse(data[:n])
		if err != nil {
			return nil, err
		}

		bytesParsed = parse

		if bytesParsed != 0 {
			moveRemainingBytesToRequestData()
		}

		if request.State == REQUEST_STATE_DONE || isEOF {
			break
		}
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {

	if bytesRead >= len(requestData) {
		allocateSpaceForRequestData()
	}
	moveReadDataToRequestData(data)

	parsedBytes, err := parseRequestLine(requestData)
	if err != nil {
		return 0, err
	}

	if parsedBytes != 0 {
		switch r.State {
		case REQUEST_STATE_INITIALIZED:
			requestLine, err := getRequestLineObjectFromRequestData(parsedBytes)

			if err != nil {
				return 0, err
			}

			r.RequestLine = *requestLine

			r.State = REQUEST_STATE_PARSING_HEADERS

			break
		case REQUEST_STATE_PARSING_HEADERS:
			_, isDone, err := r.Headers.Parse(requestData)
			if err != nil {
				return 0, err
			}

			if isDone {
				r.State = REQUEST_STATE_DONE
			}
		default:
			break
		}

	}

	return parsedBytes, err
}

func parseRequestLine(requestBytes []byte) (int, error) {
	requestString := string(requestBytes)

	if !strings.Contains(requestString, SEPARATOR) {
		return 0, nil
	}

	finishRequestLine := strings.Index(requestString, SEPARATOR)
	return len(requestBytes[:finishRequestLine]), nil

}

func allocateSpaceForRequestData() {
	aux := make([]byte, bytesRead)
	_ = copy(aux, requestData)
	requestData = make([]byte, bytesRead*2)
	_ = copy(requestData, aux)
}

func moveReadDataToRequestData(data []byte) {
	count := 0
	for i := bytesRead - len(data); i < bytesRead; i++ {
		requestData[i] = data[count]
		count += 1
	}
}

func getRequestLineObjectFromRequestData(parsedBytes int) (*RequestLine, error) {

	requestLineString := string(requestData[:parsedBytes])

	requestLineItems := strings.Split(requestLineString, " ")
	if len(requestLineItems) != 3 {
		return nil, fmt.Errorf("error: request line items malformed")
	}
	if expValue, err := regexp.MatchString("[A-Z]+", requestLineItems[0]); !expValue || err != nil {
		return nil, fmt.Errorf("error: regex failed for method")
	}
	if !strings.Contains(requestLineItems[2], "HTTP/1.1") {
		return nil, fmt.Errorf("error: only http/1.1 is supported")
	}

	return &RequestLine{
		Method:        requestLineItems[0],
		RequestTarget: requestLineItems[1],
		HttpVersion:   strings.Split(requestLineItems[2], "/")[1],
	}, nil

}

func moveRemainingBytesToRequestData() {
	remainingBytes := make([]byte, len(requestData[bytesParsed+2:bytesRead]))
	_ = copy(remainingBytes, requestData[bytesParsed+2:bytesRead])
	clear(requestData)
	bytesRead = len(remainingBytes)
	copy(requestData, remainingBytes)
}
