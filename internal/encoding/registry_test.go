package encoding

import (
	"reflect"
	"testing"

	"github.com/spf13/viper/internal/encoding/constructor"
	"github.com/spf13/viper/internal/encoding/ini"
)

type codec struct {
	v map[string]interface{}
	b []byte
}

func (c *codec) Construct() constructor.Codec {
	return &codec{}
}

func (c *codec) Encode(_ map[string]interface{}) ([]byte, error) {
	return c.b, nil
}

func (c *codec) Decode(_ []byte, v map[string]interface{}) error {
	for key, value := range c.v {
		v[key] = value
	}

	return nil
}

func TestCodecRegistry_RegisterCodec(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		registry := NewCodecRegistry("", ini.LoadOptions{})

		err := registry.RegisterCodec("myformat", &codec{})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("AlreadyRegistered", func(t *testing.T) {
		registry := NewCodecRegistry("", ini.LoadOptions{})

		err := registry.RegisterCodec("myformat", &codec{})
		if err != nil {
			t.Fatal(err)
		}

		err = registry.RegisterCodec("myformat", &codec{})
		if err != ErrCodecFormatAlreadyRegistered {
			t.Fatalf("expected ErrDecoderFormatAlreadyRegistered, got: %v", err)
		}
	})
}

func TestCodecRegistry_Decode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		registry := NewCodecRegistry("", ini.LoadOptions{})
		decoder := &codec{
			v: map[string]interface{}{
				"key": "value",
			},
		}

		err := registry.RegisterCodec("myformat", decoder)
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
		registry := NewCodecRegistry("", ini.LoadOptions{})

		v := map[string]interface{}{}

		err := registry.Decode("myformat", nil, v)
		if err != ErrCodecNotFound {
			t.Fatalf("expected ErrDecoderNotFound, got: %v", err)
		}
	})
}

func TestEncoderRegistry_Decode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		registry := NewCodecRegistry("", ini.LoadOptions{})
		encoder := &codec{
			b: []byte("key: value"),
		}

		err := registry.RegisterCodec("myformat", encoder)
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
		registry := NewCodecRegistry("", ini.LoadOptions{})

		_, err := registry.Encode("myformat", map[string]interface{}{"key": "value"})
		if err != ErrCodecNotFound {
			t.Fatalf("expected ErrEncoderNotFound, got: %v", err)
		}
	})
}
