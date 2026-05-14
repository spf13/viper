package dotenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// original form of the data.
const original = `# key-value pair
KEY=value
`

// encoded form of the data.
const encoded = `KEY=value
`

// data is Viper's internal representation.
var data = map[string]any{
	"KEY": "value",
}

func TestCodec_Encode(t *testing.T) {
	codec := Codec{}

	b, err := codec.Encode(data)
	require.NoError(t, err)

	assert.Equal(t, encoded, string(b))
}

func TestCodec_Decode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		codec := Codec{}

		v := map[string]any{}

		err := codec.Decode([]byte(original), v)
		require.NoError(t, err)

		assert.Equal(t, data, v)
	})

	t.Run("InvalidData", func(t *testing.T) {
		codec := Codec{}

		v := map[string]any{}

		err := codec.Decode([]byte(`invalid data`), v)
		require.Error(t, err)

		t.Logf("decoding failed as expected: %s", err)
	})
}

// TestCodec_Encode_StringSlice verifies that []string values are encoded as
// comma-separated values rather than Go's default bracket notation "[a b]".
//
// dotenv has no array type, so the de-facto convention is CSV (VAL=a,b,c).
// Viper's defaultDecoderConfig uses stringToWeakSliceHookFunc(",") which
// reconstructs []string from a comma-separated string during Unmarshal,
// completing the round-trip.
func TestCodec_Encode_StringSlice(t *testing.T) {
	codec := Codec{}

	b, err := codec.Encode(map[string]any{
		"HOSTS": []string{"host1:10101", "host2:10101"},
	})
	require.NoError(t, err)

	// Must be comma-separated, not bracket notation.
	assert.Equal(t, "HOSTS=host1:10101,host2:10101\n", string(b))
}
