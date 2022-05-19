package bitcask

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
	"sync"

	Utils "github.com/vinaygb665/bitcask-kv/utils"

	// Import serializer and deserializer
	S "github.com/vinaygb665/bitcask-kv/serializer"

	"github.com/schollz/progressbar/v3"
	"github.com/tidwall/btree"
)

type Storage struct {
	Mu            sync.Mutex
	IndexLock     sync.Mutex
	Threshold     int64                         // Represents the maximum threshold of the active file
	RO            bool                          // Represents if the storage is read only
	Dirname       string                        // Represents the directory name
	IndexFile     *os.File                      // Represents the index file
	ActiveFile    *os.File                      // Represents the active file
	ActiveFileId  int                           // Represents the active file id
	Keymap        map[string]*Utils.KeyInfo     // Represents the hashmap of keys
	TimestampTree *btree.Generic[Utils.KeyInfo] // Represents the btree of Keys with
	KeyTree       *btree.Generic[Utils.KeyInfo] // Represents the btree of keys
	SizeTree      *btree.Generic[Utils.KeyInfo] // Represents the btree of keys
}

func (s *Storage) Init(dirname string, ro bool, threshold int64) {

	// Basic inits create all default values
	s.Dirname = dirname
	s.Threshold = threshold
	s.Keymap = make(map[string]*Utils.KeyInfo)
	s.TimestampTree, s.KeyTree, s.SizeTree = Utils.InitializeIndexTrees()

	// Initialize index file related data
	indexFile, err := Utils.InitializeIndex(dirname)
	if err != nil {
		panic(err)
	}
	s.IndexFile = indexFile

	if s.ActiveFile == nil {

		activeFileId := Utils.InitializeDataFiles(dirname)
		if activeFileId == 0 {
			activeFileId = 1000
		} else {
			s.LoadKeys()
		}

		file, _ := os.OpenFile(s.Dirname+"/"+strconv.Itoa(activeFileId)+".dat", os.O_RDWR|os.O_CREATE, 0777)
		s.ActiveFile = file
		s.ActiveFileId = activeFileId

	}

}

func (s *Storage) Write(key string, value []byte) error {

	// Check if the active file is full
	if len(value) > Utils.MAX_VALUE_SIZE {
		return errors.New("value is too large")
	}

	err := s.UpdateIndex(key, value)
	if err != nil {
		return err
	}
	writeSuccess := make(chan bool)
	go s.__write_value(value, writeSuccess)
	if <-writeSuccess {
		return nil
	}
	return errors.New("error writing value")
	// Write index record to index file

	return nil
}

func (s *Storage) UpdateIndex(key string, value []byte) (err error) {
	s.IndexLock.Lock()
	defer s.IndexLock.Unlock()
	s.IndexFile.Seek(0, os.SEEK_END)
	offset, _ := s.ActiveFile.Seek(0, os.SEEK_END)

	indexRecord := S.CreateNewIndexRecord(key, value)
	indexRecord.FileID = s.ActiveFileId
	indexRecord.Offset = offset
	encodedIndexRecord := S.EncodeIndexRecord(indexRecord)

	_, err = s.IndexFile.WriteString(encodedIndexRecord + "\n")
	rec := &Utils.KeyInfo{
		Key:       key,
		FileID:    s.ActiveFileId,
		Offset:    offset,
		Size:      indexRecord.Size,
		Timestamp: indexRecord.Timestamp,
	}
	s.__update_index_trees(rec)
	return
	// s.__write_value(value)

}

func (s *Storage) __update_index_trees(rec *Utils.KeyInfo) {
	// Update keymap
	s.Keymap[rec.Key] = rec
	// Update index trees
	s.TimestampTree.Set(*rec)
	s.KeyTree.Set(*rec)
	s.SizeTree.Set(*rec)
}

func (s *Storage) Read(key string) ([]byte, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if keyInfo, ok := s.Keymap[key]; ok {
		// Check if fileId is the same as the active file id
		val, err := s.__read_value(keyInfo)
		if err != nil {
			return nil, err
		}
		return val, nil
	}
	return nil, errors.New("key not found")

}

func (s *Storage) __read_value(record *Utils.KeyInfo) (value []byte, err error) {

	var dataFile *os.File
	var close bool
	val := make([]byte, record.Size)

	if record.FileID == s.ActiveFileId {
		dataFile = s.ActiveFile
		close = false
	} else {
		dataFile, err = os.OpenFile(s.Dirname+"/"+strconv.Itoa(record.FileID)+".dat", os.O_RDWR, 0777)
		if err != nil {
			return nil, err
		}
		close = true
	}

	dataFile.Seek(record.Offset, io.SeekStart)
	_, err = dataFile.Read(val)
	if err != nil {
		return nil, err
	}
	// value = val
	if close {
		dataFile.Close()
	}
	return val, nil
}

func (s *Storage) Delete(key string) {
	// Do nothing
}

func (s *Storage) LoadKeys() {

	indexFileSize, err := s.IndexFile.Seek(0, io.SeekEnd)
	var bar *progressbar.ProgressBar = nil
	if err == nil {
		bar = progressbar.DefaultBytes(indexFileSize, "Indexfile read")
	}

	s.IndexFile.Seek(0, io.SeekStart)
	scanner := bufio.NewScanner(s.IndexFile)
	for scanner.Scan() {
		indexRecord := S.DecodeIndexRecord(scanner.Text())
		newRec := &Utils.KeyInfo{
			Key:       indexRecord.Key,
			FileID:    indexRecord.FileID,
			Offset:    indexRecord.Offset,
			Size:      indexRecord.Size,
			Timestamp: indexRecord.Timestamp,
		}
		// Check if the key is already in the keymap
		if entry, ok := s.Keymap[indexRecord.Key]; ok {
			// Compare timestamps and update if necessary
			if indexRecord.Timestamp > entry.Timestamp {
				s.__update_index_trees(newRec)
			}
		} else {
			s.__update_index_trees(newRec)
		}

		// Update bar with current offset of indexfile
		if bar != nil {
			curOffset, _ := s.IndexFile.Seek(0, io.SeekCurrent)
			bar.Set(int(curOffset))
		}
	}
}

func (s *Storage) __write_value(value []byte, success chan bool) {
	// Write the value to the active file
	s.Mu.Lock()
	defer s.Mu.Unlock()
	activeFileSize, _ := s.ActiveFile.Seek(0, os.SEEK_END)
	if activeFileSize > s.Threshold {
		// Create a new file
		s.ActiveFile.Close()
		s.ActiveFileId++
		s.ActiveFile, _ = os.OpenFile(s.Dirname+"/"+strconv.Itoa(s.ActiveFileId)+".dat", os.O_RDWR|os.O_CREATE, 0777)
		s.ActiveFile.Seek(0, os.SEEK_END)
	}

	_, err := s.ActiveFile.Write(value)
	if err != nil {
		panic(err)
	}
	success <- true
}

func (s *Storage) Scankeys(request *Utils.ScanKeysRequest) (result *Utils.ScanKeysResponse) {
	// Check if scan request is based on timestamp or key or size
	// If timestamp, scan the timestamp tree
	// If key, scan the key tree
	// If size, scan the size tree

	if request.OlderThan != 0 {
		result = s.__scan_timestamp(request)
		return
	} else if request.KeyGreaterThan != "" {
		result = s.__scan_keys(request)
		return
	}

	return
}

func (s *Storage) __scan_timestamp(request *Utils.ScanKeysRequest) (result *Utils.ScanKeysResponse) {
	// Scan the timestamp tree
	// Create a new ScanKeysResponse
	// Iterate over the tree and add all keys to the ScanKeysResponse
	// Return the ScanKeysResponse
	resp := &Utils.ScanKeysResponse{}
	key := &Utils.KeyInfo{
		Timestamp: request.OlderThan,
	}
	s.TimestampTree.Ascend(*key, func(item Utils.KeyInfo) bool {
		resp.Keys = append(resp.Keys, item)
		return true
	})

	return resp

}

func (s *Storage) __scan_keys(request *Utils.ScanKeysRequest) (result *Utils.ScanKeysResponse) {
	// Scan the key tree
	// Create a new ScanKeysResponse
	// Iterate over the tree and add all keys to the ScanKeysResponse
	// Return the ScanKeysResponse
	resp := &Utils.ScanKeysResponse{}
	key := &Utils.KeyInfo{
		Key: request.KeyGreaterThan,
	}
	s.KeyTree.Ascend(*key, func(item Utils.KeyInfo) bool {
		resp.Keys = append(resp.Keys, item)
		return true
	})

	return resp
}
