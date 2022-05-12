package serializer

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"hash/crc32"
	"time"
)

type KVEntry struct {
	Checksum  uint32
	Timestamp int64
	KeySize   int32
	ValueSize int32
	Key       string
	Value     []byte
}

type IndexRecord struct {
	Checksum  uint32
	Key       string
	FileID    int
	Offset    int64
	Size      int
	Timestamp int64
}

func Serialize(key string, value []byte) (int64, string) {
	// Generate CRC32 checksum for the value
	crc := crc32.Checksum(value, crc32.MakeTable(crc32.IEEE))
	// Timestamp
	timestamp := time.Now().Unix()
	kv := &KVEntry{
		Checksum:  crc,
		Timestamp: timestamp,
		KeySize:   int32(len(key)),
		ValueSize: int32(len(value)),
		Key:       key,
	}
	// Serialize the KVEntry
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(kv)
	if err != nil {
		panic(err)
	}
	serialized := base64.StdEncoding.EncodeToString(buf.Bytes())

	return timestamp, serialized
}

func Deserialize(serialized string, value_needed bool) (timestamp int64, key string, value []byte, success bool) {

	// Decode the serialized string
	decoded, err := base64.StdEncoding.DecodeString(serialized)
	if err != nil {
		return 0, "", nil, false
	}
	// Deserialize the KVEntry
	buf := bytes.NewBuffer(decoded)
	kv := &KVEntry{}
	err = gob.NewDecoder(buf).Decode(kv)
	if err != nil {
		return 0, "", nil, false
	}

	timestamp = kv.Timestamp
	key = kv.Key

	if value_needed {
		// Check if the checksum is valid
		if crc32.Checksum(kv.Value, crc32.MakeTable(crc32.IEEE)) != kv.Checksum {
			return 0, "", nil, false
		}

		value = kv.Value
		success = true
		return
	}
	success = true
	return

}

func CreateNewIndexRecord(key string, value []byte) *IndexRecord {
	// Generate CRC32 checksum for the value
	crc := crc32.Checksum(value, crc32.MakeTable(crc32.IEEE))
	// Timestamp
	timestamp := time.Now().Unix()
	kvRecord := &IndexRecord{
		Checksum:  crc,
		Key:       key,
		FileID:    0,
		Offset:    0,
		Size:      len(value),
		Timestamp: timestamp,
	}
	return kvRecord
}

func EncodeIndexRecord(record *IndexRecord) string {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(record)
	if err != nil {
		panic(err)
	}
	serialized := base64.StdEncoding.EncodeToString(buf.Bytes())
	return serialized
}

func DecodeIndexRecord(serialized string) *IndexRecord {
	// Decode the serialized string
	decoded, err := base64.StdEncoding.DecodeString(serialized)
	if err != nil {
		return nil
	}
	// Deserialize the KVEntry
	buf := bytes.NewBuffer(decoded)
	record := &IndexRecord{}
	err = gob.NewDecoder(buf).Decode(record)
	if err != nil {
		return nil
	}
	return record
}
