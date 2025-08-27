package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.OpenFile("messages.txt", os.O_RDONLY, 0666)
	if err != nil {
		return
	}

	defer file.Close()

	ch := getLinesChannel(file)

	for msg := range ch {
		fmt.Printf("read: %s\n", msg)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer f.Close()

		byteChunk := 8
		var currentLine strings.Builder
		for {
			b := make([]byte, byteChunk)
			_, err := f.Read(b)
			if err != nil {
				break
			}

			currentString := string(b)
			newLineTokens := strings.Split(currentString, "\n")

			if len(newLineTokens) > 1 {
				currentLine.WriteString(newLineTokens[0])
				line := currentLine.String()

				ch <- line
				currentLine.Reset()

				for i := 1; i < len(newLineTokens); i++ {
					if i == len(newLineTokens)-1 {
						currentLine.WriteString(newLineTokens[i])
						break
					}

					ch <- newLineTokens[i]
				}
			} else {
				currentLine.WriteString(currentString)
			}

		}
	}()

	return ch
}
