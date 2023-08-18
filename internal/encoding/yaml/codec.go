package yaml

import (
	"github.com/spf13/viper/internal/encoding/codec"
	"gopkg.in/yaml.v3"
)

// Codec implements the encoding.Codec interface for YAML encoding.
type Codec struct{}

func New(_ ...interface{}) codec.Codec {
	return &Codec{}
}

func (*Codec) Encode(v map[string]interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (*Codec) Decode(b []byte, v map[string]interface{}) error {
	return yaml.Unmarshal(b, &v)
}
