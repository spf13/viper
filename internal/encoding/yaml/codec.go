package yaml

import "go.yaml.in/yaml/v3"

// Codec implements the encoding.Encoder and encoding.Decoder interfaces for YAML encoding.
type Codec struct{}

// Encode encodes a map[string]any into a YAML byte slice.
func (Codec) Encode(v map[string]any) ([]byte, error) {
	return yaml.Marshal(v)
}

// Decode decodes a YAML byte slice into a map[string]any.
func (Codec) Decode(b []byte, v map[string]any) error {
	return yaml.Unmarshal(b, &v)
}
