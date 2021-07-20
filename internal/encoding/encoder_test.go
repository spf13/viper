package encoding

import (
	"testing"
)

type encoder struct {
	b []byte
}

func (e encoder) Encode(_ map[string]interface{}) ([]byte, error) {
	return e.b, nil
}

func TestEncoderRegistry_RegisterEncoder(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		registry := NewEncoderRegistry()

		err := registry.RegisterEncoder("myformat", encoder{})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("AlreadyRegistered", func(t *testing.T) {
		registry := NewEncoderRegistry()

		err := registry.RegisterEncoder("myformat", encoder{})
		if err != nil {
			t.Fatal(err)
		}

		err = registry.RegisterEncoder("myformat", encoder{})
		if err != ErrEncoderFormatAlreadyRegistered {
			t.Fatalf("expected ErrEncoderFormatAlreadyRegistered, got: %v", err)
		}
	})
}

func TestEncoderRegistry_Decode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		registry := NewEncoderRegistry()
		encoder := encoder{
			b: []byte("key: value"),
		}

		err := registry.RegisterEncoder("myformat", encoder)
		if err != nil {
			t.Fatal(err)
		}

		b, err := registry.Encode("myformat", map[string]interface{}{"key": "value"})
		if err != nil {
			t.Fatal(err)
		}

		if string(b) != "key: value" {
			t.Fatalf("expected 'key: value', got: %#v", string(b))
		}
	})

	t.Run("EncoderNotFound", func(t *testing.T) {
		registry := NewEncoderRegistry()

		_, err := registry.Encode("myformat", map[string]interface{}{"key": "value"})
		if err != ErrEncoderNotFound {
			t.Fatalf("expected ErrEncoderNotFound, got: %v", err)
		}
	})
}
