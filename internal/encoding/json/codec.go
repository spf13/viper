package json

import (
	"encoding/json"

	"github.com/spf13/viper/internal/encoding/codec"
)

// Codec implements the encoding.Codec interface for JSON encoding.
type Codec struct{}

func New(_ ...interface{}) codec.Codec {
	return &Codec{}
}

func (*Codec) Encode(v map[string]interface{}) ([]byte, error) {
	// TODO: expose prefix and indent in the Codec as setting?
	return json.MarshalIndent(v, "", "  ")
}

func (*Codec) Decode(b []byte, v map[string]interface{}) error {
	return json.Unmarshal(b, &v)
}
