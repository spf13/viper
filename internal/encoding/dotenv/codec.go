package dotenv

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/subosito/gotenv"
)

const keyDelimiter = "_"

// Codec implements the encoding.Encoder and encoding.Decoder interfaces for encoding data containing environment variables
// (commonly called as dotenv format).
type Codec struct{}

// Encode encodes a map[string]any into a dotenv byte slice.
func (Codec) Encode(v map[string]any) ([]byte, error) {
	flattened := map[string]any{}

	flattened = flattenAndMergeMap(flattened, v, "", keyDelimiter)

	keys := make([]string, 0, len(flattened))

	for key := range flattened {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var buf bytes.Buffer

	for _, key := range keys {
		_, err := buf.WriteString(fmt.Sprintf("%v=%v\n", strings.ToUpper(key), encodeValue(flattened[key])))
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// needsQuoting reports whether an unquoted dotenv value would not survive a
// round trip through the parser: a "#" starts a comment, "$NAME" is expanded,
// surrounding whitespace is trimmed, a newline ends the line, and a leading
// quote character is treated as the start of a quoted value.
func needsQuoting(val string) bool {
	if val == "" {
		return false
	}
	if strings.ContainsAny(val, "#\n\r$") {
		return true
	}
	if strings.TrimLeft(val, " \t") != val || strings.TrimRight(val, " \t") != val {
		return true
	}
	return val[0] == '\'' || val[0] == '"'
}

// encodeValue renders val as a dotenv value, quoting it when an unquoted
// rendering would lose data.
//
// Values without line breaks are single-quoted: the parser takes their content
// literally (no unescaping, no variable expansion) and only treats the last
// quote as the closing one. Values spanning several lines are double-quoted, so
// the entry stays on one line, with the characters the parser interprets
// escaped.
//
// The parser treats a quote preceded by a backslash as escaped, even the
// closing one, and only checks the last quote on the line. A quoted value
// ending in a backslash is therefore followed by a comment holding a quote, so
// the line still parses and the real closing quote ends the value.
func encodeValue(val any) string {
	s := fmt.Sprintf("%v", val)
	if !needsQuoting(s) {
		return s
	}

	var quoted string
	if strings.ContainsAny(s, "\n\r") {
		quoted = doubleQuote(s)
	} else {
		quoted = "'" + s + "'"
	}

	if strings.HasSuffix(s, `\`) {
		quoted += " #" + quoted[:1] + "."
	}
	return quoted
}

func doubleQuote(s string) string {
	var buf strings.Builder
	buf.Grow(len(s) + 2)
	buf.WriteByte('"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\':
			buf.WriteString(`\\`)
		case '"':
			buf.WriteString(`\"`)
		case '$':
			buf.WriteString(`\$`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		default:
			buf.WriteByte(c)
		}
	}
	buf.WriteByte('"')
	return buf.String()
}

// Decode decodes a dotenv byte slice into a map[string]any.
func (Codec) Decode(b []byte, v map[string]any) error {
	var buf bytes.Buffer

	_, err := buf.Write(b)
	if err != nil {
		return err
	}

	env, err := gotenv.StrictParse(&buf)
	if err != nil {
		return err
	}

	for key, value := range env {
		v[key] = value
	}

	return nil
}
