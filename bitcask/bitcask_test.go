package bitcask

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	Utils "github.com/vinaygb665/bitcask-kv/utils"
)

func TestWrite(t *testing.T) {

	s := Storage{}
	rand.Seed(time.Now().UnixNano())
	s.Init("/tmp/bitcask", false, 1024*1024*1024)
	key := "dummy-key-2"
	value := "dummy-value"
	err := s.Write(key, []byte(value))
	if err != nil {
		t.Errorf("Write failed")
	}
	allKeys := s.Keymap
	if len(allKeys) <= 1 {
		t.Errorf("Keymap size is less than 1")
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
	nKeys := len(s.Keymap)
	t.Logf("Load time: %s for %d keys", loadtime, nKeys)
	nWrites := 200000
	// Array to keep track of write times
	writeTimes := make([]int64, nWrites)

	for i := 0; i < nWrites; i++ {
		key := Utils.RandStringBytes(30)
		// sizeMultiplier := int(math.Pow10(int(i / 25000)))
		randomSize := rand.Intn(100)
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
	key := "dummy-key"
	value := "dummy-value"

	err := s.Write(key, []byte(value))
	if err != nil {
		t.Errorf("Write failed")
	}
	readVal, success := s.Read(key)
	if !success {
		t.Errorf("Read failed")
	}
	if string(readVal) != value {
		t.Errorf("Read value is not correct")
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
