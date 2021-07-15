package encoding

import (
	"reflect"
	"testing"
)

type decoder struct {
	v map[string]interface{}
}

func (d decoder) Decode(_ []byte, v map[string]interface{}) error {
	for key, value := range d.v {
		v[key] = value
	}

	return nil
}

func TestDecoderRegistry_RegisterDecoder(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		registry := NewDecoderRegistry()

		err := registry.RegisterDecoder("myformat", decoder{})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("AlreadyRegistered", func(t *testing.T) {
		registry := NewDecoderRegistry()

		err := registry.RegisterDecoder("myformat", decoder{})
		if err != nil {
			t.Fatal(err)
		}

		err = registry.RegisterDecoder("myformat", decoder{})
		if err != ErrDecoderFormatAlreadyRegistered {
			t.Fatalf("expected ErrDecoderFormatAlreadyRegistered, got: %v", err)
		}
	})
}

func TestDecoderRegistry_Decode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		registry := NewDecoderRegistry()
		decoder := decoder{
			v: map[string]interface{}{
				"key": "value",
			},
		}

		err := registry.RegisterDecoder("myformat", decoder)
		if err != nil {
			t.Fatal(err)
		}

		v := map[string]interface{}{}

		err = registry.Decode("myformat", []byte("key: value"), v)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(decoder.v, v) {
			t.Fatalf("decoded value does not match the expected one\nactual:   %+v\nexpected: %+v", v, decoder.v)
		}
	})

	t.Run("DecoderNotFound", func(t *testing.T) {
		registry := NewDecoderRegistry()

		v := map[string]interface{}{}

		err := registry.Decode("myformat", nil, v)
		if err != ErrDecoderNotFound {
			t.Fatalf("expected ErrDecoderNotFound, got: %v", err)
		}
	})
}
