package bitcask_storage

import (
	"bytes"
	Utils "experiments/utils"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {

	s := Storage{}
	rand.Seed(time.Now().UnixNano())
	s.Init("/tmp/bitcask", false, 1024*1024*1024)
	allKeys := s.Keymap
	key := "key2"
	// Generate random bytes
	// randomSize := rand.Intn(1024 * 1024 * 10)
	randomSize := 6001561
	key = key + strconv.Itoa(rand.Intn(256))
	randomBytes := make([]byte, randomSize)
	t.Logf("Size in KBs: %s", strconv.Itoa(randomSize/1024))
	rand.Read(randomBytes)
	// value := []byte("value")
	err := s.Write(key, randomBytes)
	if err != nil {
		t.Errorf("Write failed")
	}

	if len(allKeys) == 0 {
		t.Errorf("Keymap size is 0")
	}

}

func TestBigWrite(t *testing.T) {
	s := Storage{}
	rand.Seed(time.Now().UnixNano())
	s.Init("/tmp/bitcask", false, 1024*1024*1024)
	allKeys := s.Keymap
	fmt.Println("Keymap size: ", len(allKeys))
	key := "key2"
	// Generate random bytes
	randomSize := 1024 * 1024 * 50
	randomBytes := make([]byte, randomSize)
	rand.Read(randomBytes)
	// value := []byte("value")
	err := s.Write(key, randomBytes)
	if err == nil {
		t.Errorf("Write should have failed")
	}

}

func TestMultipleWrites(t *testing.T) {
	s := Storage{}
	rand.Seed(time.Now().UnixNano())
	start := time.Now()
	s.Init("/tmp/bitcask", false, 1024*1024*1024)
	loadtime := time.Since(start)
	t.Logf("Load time: %s", loadtime)
	nWrites := 2000
	// Array to keep track of write times
	writeTimes := make([]int64, nWrites)

	for i := 0; i < nWrites; i++ {
		key := Utils.RandStringBytes(30)
		randomSize := rand.Intn(1024 * 200)
		randomBytes := make([]byte, randomSize)
		rand.Read(randomBytes)
		start = time.Now()
		err := s.Write(key, randomBytes)
		writeTimes = append(writeTimes, time.Since(start).Nanoseconds())
		if err != nil {
			t.Errorf("Write failed")
		}

	}

	avg := Utils.Average(writeTimes)
	t.Logf("Average write time: %v", avg)
	t.Logf("Write times: %v", writeTimes)

}

func TestWriteRead(t *testing.T) {
	s := Storage{}
	rand.Seed(time.Now().UnixNano())
	s.Init("/tmp/bitcask", false, 1024*1024*1024)
	// allKeys := s.Keymap
	key := "key2"
	// Generate random bytes
	randomSize := rand.Intn(1024 * 1024 * 10)
	randomBytes := make([]byte, randomSize)
	rand.Read(randomBytes)

	err := s.Write(key, randomBytes)
	if err != nil {
		t.Errorf("Write failed")
	}
	readVal, success := s.Read(key)
	if !success {
		t.Errorf("Read failed")
	}
	if !bytes.Equal(randomBytes, readVal) {
		t.Errorf("Read value is not equal to written value")
	}

}

func TestMultiRead(t *testing.T) {
	s := Storage{}
	rand.Seed(time.Now().UnixNano())
	s.Init("/tmp/bitcask", false, 1024*1024*1024)
	allKeys := s.Keymap

	readTimes := make([]int64, len(allKeys))
	// Iterate over all keys in the keymap
	for key := range allKeys {
		start := time.Now()
		_, success := s.Read(key)
		if !success {
			t.Errorf("Read failed")
		}
		readTimes = append(readTimes, time.Since(start).Nanoseconds())
	}

	avg := Utils.Average(readTimes)
	t.Logf("Average read time: %v", avg)

}

func BenchmarkWrites(b *testing.B) {
	var exponentialInputSizes []int
	for i := 0; i < 30; i++ {
		exponentialInputSizes = append(exponentialInputSizes, int(math.Pow(2, float64(i))))
	}

	s := Storage{}
	rand.Seed(time.Now().UnixNano())
	s.Init("/tmp/bitcask", false, 1024*1024*1024)

	for _, v := range exponentialInputSizes {
		b.Run(fmt.Sprintf("%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				key := Utils.RandStringBytes(30)
				// randomSize := rand.Intn(1024 * 200)
				randomBytes := make([]byte, v)
				rand.Read(randomBytes)
				err := s.Write(key, randomBytes)
				if err != nil {
					b.Errorf("Write failed")
				}
			}
		})
	}
}
