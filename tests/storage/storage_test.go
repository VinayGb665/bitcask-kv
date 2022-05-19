package storage_test

import (
	"math/rand"
	"os"
	"testing"
	"time"

	Bitcask "github.com/vinaygb665/bitcask-kv/bitcask"
	Utils "github.com/vinaygb665/bitcask-kv/utils"
)

const TEST_DATA_DIR = "/tmp/bitcask-kv-test"

func cleanDirectory(dirname string) {
	os.RemoveAll(dirname)
}

func GenerateKeyValuePairs(n int, ksize int, vsize int) map[string]string {
	keyValuePairs := make(map[string]string)
	for i := 0; i < n; i++ {
		key, value := GenerateRandomKV(ksize, vsize)
		keyValuePairs[key] = string(value)
	}
	return keyValuePairs
}

func GenerateRandomKV(keysize int, valuesize int) (string, []byte) {
	rand.Seed(time.Now().UnixNano())
	key := Utils.RandStringBytes(keysize)
	value := make([]byte, valuesize)
	rand.Read(value)
	return key, value
}

func TestValidWrites(t *testing.T) {

	storage := &Bitcask.Storage{}
	storage.Init(TEST_DATA_DIR, false, 1024*1024*1024)
	samplesize := 50000

	keyValuePairs := GenerateKeyValuePairs(samplesize, 10, 100)
	for key, value := range keyValuePairs {

		t.Log("Writing key: ", key, " value: ", value)
		err := storage.Write(key, []byte(value))
		if err != nil {
			t.Errorf("Error writing key %s, value %s", key, value)
			return
		}

	}

	for key, value := range keyValuePairs {
		readValue, err := storage.Read(key)
		if err != nil {
			t.Errorf("Error reading key %s, value %s", key, value)
			return
		}
		if string(readValue) != value {
			t.Errorf("Error reading key %s, value %s", key, value)
			return
		}
	}
	cleanDirectory(TEST_DATA_DIR)

}
