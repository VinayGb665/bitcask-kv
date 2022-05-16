package bitcask

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	Utils "github.com/vinaygb665/bitcask-kv/utils"

	// Import serializer and deserializer
	S "github.com/vinaygb665/bitcask-kv/serializer"

	"github.com/schollz/progressbar/v3"
)

type KeyInfo struct {
	Key       string
	FileID    int
	Offset    int64
	Size      int
	Timestamp int64
}

type Storage struct {
	Mu           sync.Mutex
	Dirname      string
	IndexFile    *os.File
	RO           bool
	Threshold    int64 // Represents the maximum threshold of the active file
	ActiveFile   *os.File
	ActiveFileId int
	// Keymap maps keys to the fileid, offset and size of the value
	Keymap map[string]*KeyInfo
}

func (s *Storage) Init(dirname string, ro bool, threshold int64) {

	// Do nothing
	s.Dirname = dirname
	s.Keymap = make(map[string]*KeyInfo)
	s.Threshold = threshold

	// Create directory if it doesn't exist

	_ = os.Mkdir(dirname, 0777)

	// Check if index file exists
	// If it doesn't exist, create it
	indexFilePath := dirname + "/" + Utils.INDEX_FILE_NAME

	if _, err := os.Stat(indexFilePath); err == nil {
		// Index file exists
		// Read the index file
		s.IndexFile, err = os.OpenFile(indexFilePath, os.O_RDWR, 0777)
		if err != nil {
			panic(err)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// Index file doesn't exist
		// Create the index file
		indexFile, err := os.Create(indexFilePath)
		if err != nil {
			panic(err)
		}
		s.IndexFile = indexFile
	} else {
		panic(err)
	}

	if s.ActiveFile == nil {
		/*
			Iterate through the directory and find the filename that has the highest fileid
			Set the active file to that file
			Set the threshold to the threshold
			Create the active file and set its name to 1000.dat

		*/
		files, err := ioutil.ReadDir(s.Dirname)
		if err != nil {
			panic(err)
		}
		var maxFileID int = 0
		for _, file := range files {
			fileID, err := strconv.Atoi(file.Name()[:len(file.Name())-4])

			if err != nil {
				continue
			}
			if fileID > maxFileID {
				maxFileID = fileID
			}
		}
		if maxFileID == 0 {
			maxFileID = 1000
		} else {
			s.LoadKeys()
		}

		file, _ := os.OpenFile(s.Dirname+"/"+strconv.Itoa(maxFileID)+".dat", os.O_RDWR|os.O_CREATE, 0777)
		s.ActiveFile = file
		s.ActiveFileId = maxFileID

	}

}

func (s *Storage) Write(key string, value []byte) error {
	// Open the active file and check if it is full/threshold is reached
	// If it is, create a new file and update the active file
	// If it is not, append the value to the active file
	// Update the keymap

	// Check if the active file is full
	if len(value) > Utils.MAX_VALUE_SIZE {
		return errors.New("value is too large")
	}

	// Write to end of file, record the SEEK_END position before writing
	offset, _ := s.ActiveFile.Seek(0, os.SEEK_END)
	indexRecord := S.CreateNewIndexRecord(key, value)
	indexRecord.FileID = s.ActiveFileId
	indexRecord.Offset = offset

	go s.__write_value(value)
	// Write index record to index file
	encodedIndexRecord := S.EncodeIndexRecord(indexRecord)
	// Seek to end of file
	s.IndexFile.Seek(0, os.SEEK_END)
	// Write index record to index file
	_, err := s.IndexFile.WriteString(encodedIndexRecord + "\n")
	if err != nil {
		return err
	}

	s.Keymap[key] = &KeyInfo{
		Key:       key,
		FileID:    s.ActiveFileId,
		Offset:    offset,
		Size:      indexRecord.Size,
		Timestamp: indexRecord.Timestamp,
	}
	return nil
}

func (s *Storage) Read(key string) ([]byte, bool) {
	/*
		Check if the key is in the keymap
		- Read the value from the active file
		- Deserialize the value
		- Return the value

	*/
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if keyInfo, ok := s.Keymap[key]; ok {
		// Check if fileId is the same as the active file id
		if keyInfo.FileID == s.ActiveFileId {
			// Read the value from the active file
			s.ActiveFile.Seek(keyInfo.Offset, io.SeekStart)
			value := make([]byte, keyInfo.Size)
			_, err := s.ActiveFile.Read(value)
			if err != nil {
				return nil, false
			}
			return value, true
		} else {
			// Read the value from the inactive file
			inactiveFile, err := os.OpenFile(s.Dirname+"/"+strconv.Itoa(keyInfo.FileID)+".dat", os.O_RDWR, 0777)
			if err != nil {
				return nil, false
			}
			inactiveFile.Seek(keyInfo.Offset, io.SeekStart)
			value := make([]byte, keyInfo.Size)
			_, err = inactiveFile.Read(value)
			if err != nil {
				return nil, false
			}
			inactiveFile.Close()
			return value, true
		}
	}
	return nil, false

}

func (s *Storage) Delete(key string) {
	// Do nothing
}

func (s *Storage) Update(key string, value []byte) {
	// Do nothing
}

func (s *Storage) LoadKeys() {
	// Read the index file
	// For each line, deserialize the index record
	// Add the index record to the keymap
	indexFileSize, err := s.IndexFile.Seek(0, io.SeekEnd)
	var bar *progressbar.ProgressBar = nil
	if err == nil {
		fmt.Print("Non empty	")
		bar = progressbar.DefaultBytes(indexFileSize, "Indexfile read")
	}

	s.IndexFile.Seek(0, io.SeekStart)
	scanner := bufio.NewScanner(s.IndexFile)
	for scanner.Scan() {
		indexRecord := S.DecodeIndexRecord(scanner.Text())

		// Check if the key is already in the keymap
		if entry, ok := s.Keymap[indexRecord.Key]; ok {
			// Compare timestamps and update if necessary
			if indexRecord.Timestamp > entry.Timestamp {
				s.Keymap[indexRecord.Key] = &KeyInfo{
					Key:       indexRecord.Key,
					FileID:    indexRecord.FileID,
					Offset:    indexRecord.Offset,
					Size:      indexRecord.Size,
					Timestamp: indexRecord.Timestamp,
				}
			}
		} else {
			s.Keymap[indexRecord.Key] = &KeyInfo{
				Key:       indexRecord.Key,
				FileID:    indexRecord.FileID,
				Offset:    indexRecord.Offset,
				Size:      indexRecord.Size,
				Timestamp: indexRecord.Timestamp,
			}
		}
		// Update bar with current offset of indexfile
		if bar != nil {
			curOffset, _ := s.IndexFile.Seek(0, io.SeekCurrent)
			bar.Set(int(curOffset))
		}
	}
}

func (s *Storage) __write_value(value []byte) {
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
}
