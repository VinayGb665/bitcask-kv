package serializer

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
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

func EncodeIndexRecord(record *IndexRecord) []byte {
	buf := bytes.Buffer{}

	// err := gob.NewEncoder(&buf).Encode(record)
	crcBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(crcBytes, record.Checksum)
	buf.Write(crcBytes)
	// keySizeBytes := make([]byte, 4)
	// binary.LittleEndian.PutUint32(keySizeBytes, uint32(len(record.Key)))
	// buf.Write(keySizeBytes)

	fileIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(fileIDBytes, uint32(record.FileID))
	buf.Write(fileIDBytes)
	offsetBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(offsetBytes, uint64(record.Offset))
	buf.Write(offsetBytes)
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, uint32(record.Size))
	buf.Write(sizeBytes)
	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(record.Timestamp))
	buf.Write(timestampBytes)
	keyBytes := []byte(record.Key)
	buf.Write(keyBytes)
	return buf.Bytes()
}

func DecodeIndexRecord(serialized []byte) *IndexRecord {
	// Decode the serialized string
	// decoded, err := base64.StdEncoding.DecodeString(serialized)
	// if err != nil {
	// return nil
	// }
	// Deserialize the KVEntry
	buf := bytes.NewBuffer(serialized)
	record := &IndexRecord{}
	crcBytes := make([]byte, 4)
	buf.Read(crcBytes)
	record.Checksum = binary.LittleEndian.Uint32(crcBytes)
	// keySizeBytes := make([]byte, 4)
	// buf.Read(keySizeBytes)
	// keySize := int32(binary.LittleEndian.Uint32(keySizeBytes))
	fileIDBytes := make([]byte, 4)
	buf.Read(fileIDBytes)
	record.FileID = int(binary.LittleEndian.Uint32(fileIDBytes))
	offsetBytes := make([]byte, 8)
	buf.Read(offsetBytes)
	record.Offset = int64(binary.LittleEndian.Uint64(offsetBytes))
	sizeBytes := make([]byte, 4)
	buf.Read(sizeBytes)
	record.Size = int(binary.LittleEndian.Uint32(sizeBytes))
	timestampBytes := make([]byte, 8)
	buf.Read(timestampBytes)
	record.Timestamp = int64(binary.LittleEndian.Uint64(timestampBytes))
	keyBytes := make([]byte, buf.Len())
	buf.Read(keyBytes)
	record.Key = string(keyBytes)
	return record

}
