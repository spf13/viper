package javaproperties

import (
	"reflect"
	"testing"
)

// original form of the data
const original = `#key-value pair
key = value
map.key = value
`

// encoded form of the data
const encoded = `key = value
map.key = value
`

// Viper's internal representation
var data = map[string]interface{}{
	"key": "value",
	"map": map[string]interface{}{
		"key": "value",
	},
}

func TestCodec_Encode(t *testing.T) {
	codec := Codec{}

	b, err := codec.Encode(data)
	if err != nil {
		t.Fatal(err)
	}

	if encoded != string(b) {
		t.Fatalf("decoded value does not match the expected one\nactual:   %#v\nexpected: %#v", string(b), encoded)
	}
}

func TestCodec_Decode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		codec := Codec{}

		v := map[string]interface{}{}

		err := codec.Decode([]byte(original), v)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(data, v) {
			t.Fatalf("decoded value does not match the expected one\nactual:   %#v\nexpected: %#v", v, data)
		}
	})

	t.Run("InvalidData", func(t *testing.T) {
		t.Skip("TODO: needs invalid data example")

		codec := Codec{}

		v := map[string]interface{}{}

		codec.Decode([]byte(``), v)

		if len(v) > 0 {
			t.Fatalf("expected map to be empty when data is invalid\nactual: %#v", v)
		}
	})
}

func TestCodec_DecodeEncode(t *testing.T) {
	codec := Codec{}

	v := map[string]interface{}{}

	err := codec.Decode([]byte(original), v)
	if err != nil {
		t.Fatal(err)
	}

	b, err := codec.Encode(data)
	if err != nil {
		t.Fatal(err)
	}

	if original != string(b) {
		t.Fatalf("encoded value does not match the original\nactual:   %#v\nexpected: %#v", string(b), original)
	}
}
