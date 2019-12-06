package encoding

import (
	"testing"
)

type encoder struct {
	b []byte
}

func (e encoder) Encode(_ interface{}) ([]byte, error) {
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
			b: []byte("encoded value"),
		}

		err := registry.RegisterEncoder("myformat", encoder)
		if err != nil {
			t.Fatal(err)
		}

		b, err := registry.Encode("myformat", "some value")
		if err != nil {
			t.Fatal(err)
		}

		if string(b) != "encoded value" {
			t.Fatalf("expected 'encoded value', got: %#v", string(b))
		}
	})

	t.Run("EncoderNotFound", func(t *testing.T) {
		registry := NewEncoderRegistry()

		_, err := registry.Encode("myformat", "some value")
		if err != ErrEncoderNotFound {
			t.Fatalf("expected ErrEncoderNotFound, got: %v", err)
		}
	})
}
