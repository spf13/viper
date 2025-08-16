package json

import (
	"encoding/json"
)

// Codec implements the encoding.Encoder and encoding.Decoder interfaces for JSON encoding.
type Codec struct{}

// Encode encodes a map[string]any into a JSON byte slice.
func (Codec) Encode(v map[string]any) ([]byte, error) {
	// TODO: expose prefix and indent in the Codec as setting?
	return json.MarshalIndent(v, "", "  ")
}

// Decode decodes a JSON byte slice into a map[string]any.
func (Codec) Decode(b []byte, v map[string]any) error {
	return json.Unmarshal(b, &v)
}
