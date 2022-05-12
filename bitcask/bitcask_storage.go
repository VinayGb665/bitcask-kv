package bitcask_storage

import (
	"bufio"
	"errors"
	Utils "experiments/utils"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	// Import serializer and deserializer
	S "experiments/serializer"
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

	activeFileSize, _ := s.ActiveFile.Seek(0, os.SEEK_END)
	if activeFileSize > s.Threshold {
		// Create a new file
		s.ActiveFile.Close()
		s.ActiveFileId++
		s.ActiveFile, _ = os.OpenFile(s.Dirname+"/"+strconv.Itoa(s.ActiveFileId)+".dat", os.O_RDWR|os.O_CREATE, 0777)
		s.ActiveFile.Seek(0, os.SEEK_END)
	}

	timestamp, writeValue := S.Serialize(key, value)

	// Write to end of file, record the SEEK_END position before writing
	offset, _ := s.ActiveFile.Seek(0, os.SEEK_END)

	// writeSize, err := s.ActiveFile.Write([]byte(writeValue))
	writeSize, err := s.ActiveFile.WriteString(writeValue)
	if err != nil {
		return err
	}
	// Write a newline to the end of the file
	_, err = s.ActiveFile.Write([]byte("\n"))
	if err != nil {
		return err
	}

	s.Keymap[key] = &KeyInfo{
		Key:       key,
		FileID:    s.ActiveFileId,
		Offset:    offset,
		Size:      writeSize,
		Timestamp: timestamp,
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
		// Read the value from the active file
		s.ActiveFile.Seek(keyInfo.Offset, io.SeekStart)
		buffer := make([]byte, keyInfo.Size)
		reader := bufio.NewReader(s.ActiveFile)
		_, err := reader.Read(buffer)
		if err != nil {
			return nil, false
		}
		// Deserialize the value
		_, _, value, success := S.Deserialize(string(buffer), true)
		if !success {
			return nil, false
		}
		return value, true
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
	// List all the .dat files in the directory
	// For each file, read the file and parse the key and value
	// Add the key and value to the keymap

	files, _ := ioutil.ReadDir(s.Dirname)
	for _, file := range files {
		if file.Name()[len(file.Name())-4:] != ".dat" {
			continue
		}
		fileID, _ := strconv.Atoi(file.Name()[:len(file.Name())-4])

		// Open the file
		// Read the file, split on newline
		// For each line, parse the key and value
		// Add the key and value to the keymap
		// Check if fileID ends with .dat
		// If it does, add the key and value to the keymap
		// If it doesn't, continue

		dataFile, err := os.OpenFile(s.Dirname+"/"+strconv.Itoa(fileID)+".dat", os.O_RDONLY, 0777)
		if err != nil {
			panic(err)
		}
		sc := bufio.NewScanner(dataFile)
		// sc.Split(Utils.ByteSplitFunc)

		// Set buffer size to utils.MAX_VALUE_SIZE
		sc.Buffer(make([]byte, Utils.MAX_VALUE_SIZE), Utils.MAX_VALUE_SIZE)
		for sc.Scan() {
			entry := sc.Text()
			// Get current offset
			offset, _ := dataFile.Seek(0, io.SeekCurrent)

			timestamp, key, _, success := S.Deserialize(entry, false)
			if success {
				// Check if the key is already in the keymap
				// If it is, check if the timestamp is newer
				// If it is, update the keymap
				// If it isn't, continue
				fileID, _ = strconv.Atoi(file.Name()[:len(file.Name())-4])

				if val, ok := s.Keymap[key]; ok {
					if val.Timestamp < timestamp {
						s.Keymap[key] = &KeyInfo{
							Key:       key,
							FileID:    fileID,
							Offset:    offset,
							Size:      len(entry),
							Timestamp: timestamp,
						}
					}
				} else {
					s.Keymap[key] = &KeyInfo{
						Key:       key,
						FileID:    fileID,
						Offset:    offset,
						Size:      len(entry),
						Timestamp: timestamp,
					}
				}

			}
		}

	}

}

func BitcaskStorage(dirname string) {
	storage := Storage{Dirname: dirname, RO: false}
	storage.Mu.Lock()
	defer storage.Mu.Unlock()
	// Check if the directory exists
	// Check if writeable

	// Create the directory if it doesn't exist
	// Create the active file

}
