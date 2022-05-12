package serializer_test

// Import serializer and deserializer
import (
	"bytes"
	S "experiments/serializer"
	"math/rand"
	"strconv"
	"testing"
)

func TestSerializer(t *testing.T) {
	// Create a key and value
	// value := []byte("value")

	key := "key2"
	// Generate random bytes
	randomSize := rand.Intn(1024 * 1024 * 20)
	key = key + strconv.Itoa(rand.Intn(1024))
	randomBytes := make([]byte, randomSize)
	rand.Read(randomBytes)
	// Serialize the key and value
	_, serialized := S.Serialize(key, randomBytes)
	// Deserialize the serialized key and value
	_, key2, value2, ok := S.Deserialize(serialized, true)
	// Check if the deserialized key and value are the same as the original
	if key != key2 || !bytes.Equal(randomBytes, value2) || !ok {
		t.Errorf("Serializer/Deserializer failed")
	}
}