package cue

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
)

// Codec implements the encoding.Encoder and encoding.Decoder interfaces for CUE encoding.
type Codec struct{}

func (Codec) Encode(v interface{}) ([]byte, error) {
	context := cuecontext.New()
	val := context.Encode(v)
	if val.Err() != nil {
		return nil, val.Err()
	}
	return format.Node(val.Syntax(cue.ResolveReferences(true)))
}

func (Codec) Decode(b []byte, v interface{}) error {
	context := cuecontext.New()
	val := context.CompileBytes(b)
	if val.Err() != nil {
		return val.Err()
	}
	return val.Decode(v)
}
