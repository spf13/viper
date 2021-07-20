package ini

import (
	"reflect"
	"testing"
)

// original form of the data
const original = `; key-value pair
key=value ; key-value pair

# map
[map] # map
key=%(key)s

`

// encoded form of the data
const encoded = `key=value

[map]
key=value

`

// decoded form of the data
//
// in case of INI it's slightly different from Viper's internal representation
// (eg. top level keys land in a section called default)
var decoded = map[string]interface{}{
	"DEFAULT": map[string]interface{}{
		"key": "value",
	},
	"map": map[string]interface{}{
		"key": "value",
	},
}

// Viper's internal representation
var data = map[string]interface{}{
	"key": "value",
	"map": map[string]interface{}{
		"key": "value",
	},
}

func TestCodec_Encode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		codec := Codec{}

		b, err := codec.Encode(data)
		if err != nil {
			t.Fatal(err)
		}

		if encoded != string(b) {
			t.Fatalf("decoded value does not match the expected one\nactual:   %#v\nexpected: %#v", string(b), encoded)
		}
	})

	t.Run("Default", func(t *testing.T) {
		codec := Codec{}

		data := map[string]interface{}{
			"default": map[string]interface{}{
				"key": "value",
			},
			"map": map[string]interface{}{
				"key": "value",
			},
		}

		b, err := codec.Encode(data)
		if err != nil {
			t.Fatal(err)
		}

		if encoded != string(b) {
			t.Fatalf("decoded value does not match the expected one\nactual:   %#v\nexpected: %#v", string(b), encoded)
		}
	})
}

func TestCodec_Decode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		codec := Codec{}

		v := map[string]interface{}{}

		err := codec.Decode([]byte(original), v)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(decoded, v) {
			t.Fatalf("decoded value does not match the expected one\nactual:   %#v\nexpected: %#v", v, decoded)
		}
	})

	t.Run("InvalidData", func(t *testing.T) {
		codec := Codec{}

		v := map[string]interface{}{}

		err := codec.Decode([]byte(`invalid data`), v)
		if err == nil {
			t.Fatal("expected decoding to fail")
		}

		t.Logf("decoding failed as expected: %s", err)
	})
}
