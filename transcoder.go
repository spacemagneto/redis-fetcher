package fetcher

import (
	"github.com/goccy/go-json"
)

// Transcoder defines the contract for bidirectional conversion between a value of type T
// and its string representation. Users may implement custom transcoders (e.g. protobuf,
// msgpack, custom compression, etc.) to control exactly how data is serialized and stored.
// The interface is intentionally minimal and string-based because Redis stores values as strings.
type Transcoder[T any] interface {
	// Encode converts a value of type T into a string suitable for storage in Redis.
	// The implementation fully controls the format, encoding, and optional compression.
	Encode(T) (string, error)

	// Decode reconstructs a value of type T from the string previously produced by Encode.
	// It must perfectly reverse the Encode operation for the same transcoder instance.
	Decode(string) (T, error)
}

// defaultTranscoder is the built-in transcoder used when the user does not provide a custom one.
// It performs straightforward JSON serialization with no additional compression.
// This makes payloads human-readable and is perfect for development, debugging,
// or situations where size is not a critical concern.
// Users who need smaller storage footprint or a different format should supply their own transcoder.
type defaultTranscoder[T any] struct{}

// Encode method converts the provided value into a JSON string representation.
// Method serializes the input value into bytes using JSON encoding and then converts those bytes into a string.
// Any error produced during the serialization process is returned to the caller for handling.
// This method ensures that values can be safely stored in systems that expect string data.
func (defaultTranscoder[T]) Encode(src T) (string, error) {
	var bytes []byte
	var err error

	bytes, err = json.Marshal(src)

	return string(bytes), err
}

// Decode method reconstructs a value of the original type from its string representation.
// Method converts the string back into bytes and uses JSON decoding to populate the target value.
// Any error encountered during decoding is returned to the caller for proper handling.
// This method ensures that stored string data can be converted back into a usable typed value.
func (defaultTranscoder[T]) Decode(src string) (T, error) {
	var entry T

	if err := json.Unmarshal([]byte(src), &entry); err != nil {
		return entry, err
	}

	return entry, nil
}
