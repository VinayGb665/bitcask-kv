package utils

import (
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/tidwall/btree"
)

const INDEX_FILE_NAME = "inde-x"
const EndString = "#@#"

const MAX_VALUE_SIZE = 1024 * 1024 * 20
const MAX_KEY_SIZE = 1024 * 1024

type GetRequest struct {
	Key string
}
type GetResponse struct {
	Value   []byte
	Success bool
}

type SetRequest struct {
	Key   string
	Value []byte
}
type SetResponse struct {
	Success bool
}

type KeyInfo struct {
	Key       string
	FileID    int
	Offset    int64
	Size      int
	Timestamp int64
}

type ScanKeysRequest struct {
	OlderThan       int64
	Limit           int
	ValueBiggerThan int64
	KeyGreaterThan  string
}

type ScanKeysResponse struct {
	Keys  []KeyInfo
	Error error
	Count int
}

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

func Contains(element string, list []string) bool {
	for _, e := range list {
		if e == element {
			return true
		}
	}
	return false
}

func InitializeIndex(dirname string) (file *os.File, err error) {
	_ = os.Mkdir(dirname, 0777)

	// Check if index file exists
	// If it doesn't exist, create it
	indexFilePath := dirname + "/" + INDEX_FILE_NAME

	if _, err = os.Stat(indexFilePath); err == nil {
		// Index file exists
		// Read the index file
		file, err = os.OpenFile(indexFilePath, os.O_RDWR, 0777)
		return
	} else if errors.Is(err, os.ErrNotExist) {
		// Index file doesn't exist
		// Create the index file
		file, err = os.Create(indexFilePath)
		return
	}
	return
}

func InitializeDataFiles(dirname string) (activeFileId int) {
	var maxFileID int = 0
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fileID, err := strconv.Atoi(file.Name()[:len(file.Name())-4])

		if err != nil {
			continue
		}
		if fileID > maxFileID {
			maxFileID = fileID
		}
	}
	return maxFileID
}

func byTimeStamp(a, b KeyInfo) bool {
	return a.Timestamp < b.Timestamp
}

func byKey(a, b KeyInfo) bool {
	return a.Key < b.Key
}

func byValSize(a, b KeyInfo) bool {
	return a.Size < b.Size
}

func InitializeIndexTrees() (
	timestamp_tree *btree.Generic[KeyInfo],
	keys_tree *btree.Generic[KeyInfo],
	valuesize_tree *btree.Generic[KeyInfo]) {

	timestamp_tree = btree.NewGeneric[KeyInfo](byTimeStamp)
	keys_tree = btree.NewGeneric[KeyInfo](byKey)
	valuesize_tree = btree.NewGeneric[KeyInfo](byValSize)
	return

}
