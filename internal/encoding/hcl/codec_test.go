package hcl

import (
	"reflect"
	"testing"
)

// original form of the data
const original = `# key-value pair
"key" = "value"

// list
"list" = ["item1", "item2", "item3"]

/* map */
"map" = {
  "key" = "value"
}

/*
nested map
*/
"nested_map" "map" {
  "key" = "value"

  "list" = ["item1", "item2", "item3"]
}`

// encoded form of the data
const encoded = `"key" = "value"

"list" = ["item1", "item2", "item3"]

"map" = {
  "key" = "value"
}

"nested_map" "map" {
  "key" = "value"

  "list" = ["item1", "item2", "item3"]
}`

// decoded form of the data
//
// in case of HCL it's slightly different from Viper's internal representation
// (eg. map is decoded into a list of maps)
var decoded = map[string]interface{}{
	"key": "value",
	"list": []interface{}{
		"item1",
		"item2",
		"item3",
	},
	"map": []map[string]interface{}{
		{
			"key": "value",
		},
	},
	"nested_map": []map[string]interface{}{
		{
			"map": []map[string]interface{}{
				{
					"key": "value",
					"list": []interface{}{
						"item1",
						"item2",
						"item3",
					},
				},
			},
		},
	},
}

// Viper's internal representation
var data = map[string]interface{}{
	"key": "value",
	"list": []interface{}{
		"item1",
		"item2",
		"item3",
	},
	"map": map[string]interface{}{
		"key": "value",
	},
	"nested_map": map[string]interface{}{
		"map": map[string]interface{}{
			"key": "value",
			"list": []interface{}{
				"item1",
				"item2",
				"item3",
			},
		},
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
