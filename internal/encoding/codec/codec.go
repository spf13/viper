package codec

type Codec interface {
	// Decode decodes the contents of b into v.
	// It's primarily used for decoding contents of a file into a map[string]interface{}.
	Decode(b []byte, v map[string]interface{}) error

	// Encode encodes the contents of v into a byte representation.
	// It's primarily used for encoding a map[string]interface{} into a file format.
	Encode(v map[string]interface{}) ([]byte, error)
}
