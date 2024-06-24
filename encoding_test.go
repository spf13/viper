package viper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type codec struct{}

func (codec) Encode(_ map[string]any) ([]byte, error) {
	return nil, nil
}

func (codec) Decode(_ []byte, _ map[string]any) error {
	return nil
}

func TestDefaultCodecRegistry(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		registry := NewCodecRegistry()

		c := codec{}

		err := registry.RegisterCodec("myformat", c)
		require.NoError(t, err)

		encoder, err := registry.Encoder("myformat")
		require.NoError(t, err)

		assert.Equal(t, c, encoder)

		decoder, err := registry.Decoder("myformat")
		require.NoError(t, err)

		assert.Equal(t, c, decoder)
	})
}
