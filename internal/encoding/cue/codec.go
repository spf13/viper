package cue

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue/cuecontext"
)

// Codec implements the encoding.Encoder and encoding.Decoder interfaces for CUE encoding.
type Codec struct{}

func (Codec) Encode(v interface{}) ([]byte, error) {
	context := cuecontext.New()
	val := context.Encode(v)
	if val.Err() != nil {
		return nil, val.Err()
	}
	valStr := fmt.Sprintf("%v", val) // uses cue.Value.Formatter to export to string
	return []byte(valStr), nil
}

func (Codec) Decode(b []byte, v interface{}) error {
	context := cuecontext.New()
	val := context.CompileBytes(b)
	fmt.Printf("%v", val)
	if val.Err() != nil {
		return val.Err()
	}
	// marshal to json so it can be unmarshaled to v
	jsonVal, _ := val.MarshalJSON()
	return json.Unmarshal(jsonVal, v)
}
