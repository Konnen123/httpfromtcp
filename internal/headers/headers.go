package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

var (
	ERROR_MALFORMED_HEADER = fmt.Errorf("error: malformed header")
	separator              = []byte("\r\n")
	fieldNameRegex         = "^[a-zA-Z0-9!#$%&'*+.^_`|~-]+$"
)

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	bytesRead := 0

	if bytes.HasPrefix(data, separator) {
		return 0, true, nil
	}

	for {
		separatorIndex := bytes.Index(data[bytesRead:], separator)
		if separatorIndex == -1 || separatorIndex == 0 {
			break
		}

		headerBytes := data[bytesRead : bytesRead+separatorIndex]

		bytesRead += len(headerBytes) + len(separator)

		headerBytes = bytes.Trim(headerBytes, " ")
		headerTokens := bytes.SplitN(headerBytes, []byte(":"), 2)

		if len(headerTokens) != 2 {
			return 0, false, ERROR_MALFORMED_HEADER
		}

		fieldName := string(headerTokens[0])
		fieldValue := strings.Trim(string(headerTokens[1]), " ")

		if strings.HasSuffix(fieldName, " ") {
			return 0, false, ERROR_MALFORMED_HEADER
		}

		re := regexp.MustCompile(fieldNameRegex)
		if !re.MatchString(fieldName) {
			return 0, false, fmt.Errorf("error: regex not matching")
		}

		fieldName = strings.ToLower(fieldName)

		if h[fieldName] != "" {
			h[fieldName] += ", " + fieldValue
		} else {
			h[fieldName] = fieldValue
		}
	}

	return bytesRead, false, nil

}

func (h Headers) GetHeaderValue(name string) (string, error) {
	value, ok := h[name]
	if !ok {
		return "", fmt.Errorf("error: name does not exist")
	}

	return value, nil
}
