package javaproperties

import (
	"bytes"
	"sort"
	"strings"

	"github.com/magiconair/properties"
	"github.com/spf13/cast"
)

// Codec implements the encoding.Encoder and encoding.Decoder interfaces for Java properties encoding.
type Codec struct {
	KeyDelimiter string
}

func (c Codec) Encode(v map[string]interface{}) ([]byte, error) {
	p := properties.NewProperties()

	flattened := map[string]interface{}{}

	flattened = flattenAndMergeMap(flattened, v, "", c.keyDelimiter())

	keys := make([]string, 0, len(flattened))

	for key := range flattened {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		_, _, err := p.Set(key, cast.ToString(flattened[key]))
		if err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer

	_, err := p.WriteComment(&buf, "#", properties.UTF8)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c Codec) Decode(b []byte, v map[string]interface{}) error {
	p, err := properties.Load(b, properties.UTF8)
	if err != nil {
		return err
	}

	for _, key := range p.Keys() {
		// ignore existence check: we know it's there
		value, _ := p.Get(key)

		// recursively build nested maps
		path := strings.Split(key, c.keyDelimiter())
		lastKey := strings.ToLower(path[len(path)-1])
		deepestMap := deepSearch(v, path[0:len(path)-1])

		// set innermost value
		deepestMap[lastKey] = value
	}

	return nil
}

func (c Codec) keyDelimiter() string {
	if c.KeyDelimiter == "" {
		return "."
	}

	return c.KeyDelimiter
}
