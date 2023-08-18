package hcl

import (
	"bytes"
	"encoding/json"

	"github.com/spf13/viper/internal/encoding/codec"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/printer"
)

// Codec implements the encoding.Codec interface for HCL encoding.
// TODO: add printer config to the codec?
type Codec struct{}

func New(args ...interface{}) codec.Codec {
	return &Codec{}
}

func (*Codec) Encode(v map[string]interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// TODO: use printer.Format? Is the trailing newline an issue?

	ast, err := hcl.Parse(string(b))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	err = printer.Fprint(&buf, ast.Node)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (*Codec) Decode(b []byte, v map[string]interface{}) error {
	return hcl.Unmarshal(b, &v)
}
