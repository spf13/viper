package viper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeByDefaultUnsignedWidths(t *testing.T) {
	for _, tc := range []struct {
		name         string
		defaultValue any
		input        string
		want         any
	}{
		{"uint8", uint8(1), "42", uint8(42)},
		{"uint8_max", uint8(1), "255", uint8(255)},
		{"uint16", uint16(1), "42", uint16(42)},
		{"uint16_max", uint16(1), "65535", uint16(65535)},
		{"uint", uint(1), "42", uint(42)},
		{"uint32", uint32(1), "42", uint32(42)},
		{"uint64", uint64(1), "42", uint64(42)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, source := range []string{"override", "env"} {
				t.Run(source, func(t *testing.T) {
					v := New()
					v.SetDefault("service.number", tc.defaultValue)
					v.SetTypeByDefaultValue(true)
					assert.Equal(t, tc.defaultValue, v.Get("service.number"))

					if source == "env" {
						t.Setenv("VIPER_TEST_UNSIGNED_NUMBER", tc.input)
						require.NoError(t, v.BindEnv("service.number", "VIPER_TEST_UNSIGNED_NUMBER"))
					} else {
						v.Set("service.number", tc.input)
					}

					assert.Equal(t, tc.want, v.Get("SERVICE.NUMBER"))
					v.SetTypeByDefaultValue(false)
					assert.Equal(t, tc.input, v.Get("service.number"))
				})
			}
		})
	}
}
