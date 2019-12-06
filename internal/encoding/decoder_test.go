package encoding

import (
	"testing"
)

type decoder struct {
	v interface{}
}

func (d decoder) Decode(_ []byte, v interface{}) error {
	rv := v.(*string)
	*rv = d.v.(string)

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
			v: "decoded value",
		}

		err := registry.RegisterDecoder("myformat", decoder)
		if err != nil {
			t.Fatal(err)
		}

		var v string

		err = registry.Decode("myformat", []byte("some value"), &v)
		if err != nil {
			t.Fatal(err)
		}

		if v != "decoded value" {
			t.Fatalf("expected 'decoded value', got: %#v", v)
		}
	})

	t.Run("DecoderNotFound", func(t *testing.T) {
		registry := NewDecoderRegistry()

		var v string

		err := registry.Decode("myformat", []byte("some value"), &v)
		if err != ErrDecoderNotFound {
			t.Fatalf("expected ErrDecoderNotFound, got: %v", err)
		}
	})
}
