package utils

import (
	"math/rand"
	"strings"
)

const INDEX_FILE_NAME = "inde-x"
const EndString = "#@#"

const MAX_VALUE_SIZE = 1024 * 1024 * 20
const MAX_KEY_SIZE = 1024 * 1024

func ByteSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {

	// Return nothing if at end of file and no data passed
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Find the index of the input of a newline followed by a
	// pound sign.
	if i := strings.Index(string(data), EndString); i >= 0 {
		return i + 1, data[0:i], nil
	}

	// If at end of file with data return the data
	if atEOF {
		return len(data), data, nil
	}

	return
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func Average(elements []int64) int64 {
	var sum int64
	for _, element := range elements {
		sum += element
	}
	return sum / int64(len(elements))
}
