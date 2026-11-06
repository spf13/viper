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

// TestCodec_RoundTrip covers values an unquoted dotenv entry cannot represent:
// the parser starts a comment on "#", expands "$NAME", trims surrounding
// whitespace and stops reading at a line break, so those values must be quoted
// on encode to survive a second ReadConfig.
func TestCodec_RoundTrip(t *testing.T) {
	// Make accidental variable expansion visible.
	t.Setenv("NAME", "EXPANDED")
	t.Setenv("HOME", "/home/expander")

	tests := []struct {
		name  string
		value string
	}{
		{"Plain", "value"},
		{"Empty", ""},
		{"InnerSpace", "hello world"},
		{"LeadingSpace", " leading"},
		{"TrailingSpace", "trailing "},
		{"Comment", "a#b"},
		{"OnlyComment", "#"},
		{"Variable", "$NAME"},
		{"BracedVariable", "${NAME}"},
		{"EmbeddedVariable", "a$NAMEb"},
		{"SingleQuote", "don't"},
		{"SurroundingSingleQuotes", "'quoted'"},
		{"LeadingSingleQuote", `'quoted`},
		{"TrailingSingleQuote", `quoted'`},
		{"DoubleQuotes", `say "hi"`},
		{"Backslash", `C:\path\to\node`},
		{"Newline", "line1\nline2"},
		{"CarriageReturn", "line1\r\nline2"},
		{"MultilineWithQuote", "line1\nit's #1 and $NAME"},
		{"MultilineWithBackslash", "a\nb\\c"},
		{"TrailingBackslash", `C:\path\`},
		{"QuotedTrailingBackslash", `a#b\`},
		{"QuotedTrailingBackslashWithQuote", `it's #1\`},
		{"VariableTrailingBackslash", `$NAME\`},
		{"MultilineTrailingBackslash", "a\nb\\"},
	}

	codec := Codec{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := codec.Encode(map[string]any{"key": tt.value})
			require.NoError(t, err)

			v := map[string]any{}

			require.NoError(t, codec.Decode(b, v))

			assert.Equal(t, tt.value, v["KEY"], "encoded as %q", string(b))
		})
	}
}

func TestCodec_EncodeQuotesOnlyWhenNeeded(t *testing.T) {
	codec := Codec{}

	b, err := codec.Encode(map[string]any{
		"plain":  "value",
		"padded": "value ",
		"secret": "a#b",
		"multi":  "line1\nline2",
		"path":   `a#b\`,
	})
	require.NoError(t, err)

	assert.Equal(t, `MULTI="line1\nline2"
PADDED='value '
PATH='a#b\' #'.
PLAIN=value
SECRET='a#b'
`, string(b))
}
