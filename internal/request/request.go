package request

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var (
	requestData              []byte
	bytesRead                int
	bytesParsed              int
	contentLengthHeaderValue int
)

const (
	REQUEST_STATE_INITIALIZED     = 0
	REQUEST_STATE_PARSING_HEADERS = 1
	REQUEST_STATE_PARSING_BODY    = 2
	REQUEST_STATE_DONE            = 3
	SEPARATOR                     = "\r\n"
)

func RequestFromReader(reader io.Reader) (*Request, error) {

	bytesRead = 0
	bytesParsed = 0
	requestData = make([]byte, 8)
	contentLengthHeaderValue = -1
	request := &Request{
		State:   REQUEST_STATE_INITIALIZED,
		Headers: headers.Headers{},
		Body:    make([]byte, 0),
	}

	isEOF := false

	for {
		data := make([]byte, 8)
		n, err := reader.Read(data)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			//isEOF = true
		}

		bytesRead += n

		parse, err := request.parse(data[:n])
		if err != nil {
			return nil, err
		}

		bytesParsed = parse
		if parse != 0 {
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

	if parsedBytes != 0 || r.State == REQUEST_STATE_PARSING_BODY {
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
			bytesParsedHeader, isDone, err := r.Headers.Parse(requestData)
			if err != nil {
				return 0, err
			}
			parsedBytes = bytesParsedHeader

			if isDone || strings.Contains(string(requestData), fmt.Sprintf("%s%s", SEPARATOR, SEPARATOR)) {
				r.State = REQUEST_STATE_PARSING_BODY

				value, err := r.Headers.GetHeaderValue("content-length")
				if err != nil {
					r.State = REQUEST_STATE_DONE
					break
				}

				contentLengthHeaderValue, err = strconv.Atoi(value)
				if err != nil {
					return 0, err
				}

			}
			break
		case REQUEST_STATE_PARSING_BODY:
			r.Body = append(r.Body, requestData[:bytesRead]...)
			parsedBytes = bytesRead

			if len(r.Body) > contentLengthHeaderValue {
				return 0, fmt.Errorf("error: body length is larger than content lenght")
			}
			if len(r.Body) == contentLengthHeaderValue {
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
	return len(requestBytes[:finishRequestLine]) + len(SEPARATOR), nil

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
	remainingBytes := make([]byte, len(requestData[bytesParsed:bytesRead]))
	_ = copy(remainingBytes, requestData[bytesParsed:bytesRead])
	clear(requestData)
	bytesRead = len(remainingBytes)
	copy(requestData, remainingBytes)
}

func (r *Request) isRequestBodySizeEqualToContentLength() (bool, error) {
	value, err := r.Headers.GetHeaderValue("content-length")
	if err != nil && len(r.Body) == 0 {
		return true, nil
	}

	contentLength := -1
	if err == nil {
		contentLength, err = strconv.Atoi(value)
		if err != nil {
			return false, err
		}
	}

	if contentLength != -1 && contentLength != len(r.Body) {
		return false, fmt.Errorf("error: content-length is not equal to request body length")
	}

	return true, nil
}
