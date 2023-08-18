package toml

import (
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper/internal/encoding/codec"
)

// Codec implements the encoding.Codec interface for TOML encoding.
type Codec struct{}

func New(_ ...interface{}) codec.Codec {
	return &Codec{}
}

func (*Codec) Encode(v map[string]interface{}) ([]byte, error) {
	return toml.Marshal(v)
}

func (*Codec) Decode(b []byte, v map[string]interface{}) error {
	return toml.Unmarshal(b, &v)
}
