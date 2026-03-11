package viper

import (
	"bytes"
	"strings"
	"testing"
)

func runConfigFuzz(tb testing.TB, data []byte, configType string) {
	if len(data) > 1<<20 {
		return
	}

	v := New()
	v.SetConfigType(configType)

	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return
	}

	allKeys := v.AllKeys()
	for _, key := range allKeys {
		// Exercise various getters to maximize code coverage
		_ = v.Get(key)
		_ = v.GetString(key)
		_ = v.GetInt(key)
		_ = v.GetBool(key)
		_ = v.GetStringSlice(key)
		_ = v.IsSet(key)
	}
	v.AllSettings()
	_ = v.Unmarshal(&map[string]any{})
}

func FuzzConfigYAML(f *testing.F) {
	f.Add([]byte(`key: value`))
	f.Add([]byte(`nested: { key: value }`))
	f.Add([]byte(`key: null`))
	f.Add([]byte(strings.Repeat("a: b\n", 100)))
	f.Add([]byte(``))
	f.Add([]byte(`---`))
	f.Add([]byte(`~`))

	f.Fuzz(func(t *testing.T, data []byte) {
		runConfigFuzz(t, data, "yaml")
	})
}

func FuzzConfigJSON(f *testing.F) {
	f.Add([]byte(`{"key": "value"}`))
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		runConfigFuzz(t, data, "json")
	})
}

func FuzzConfigTOML(f *testing.F) {
	f.Add([]byte(`key = "value"`))
	f.Add([]byte(``))
	f.Fuzz(func(t *testing.T, data []byte) {
		runConfigFuzz(t, data, "toml")
	})
}

func FuzzConfigPriority(f *testing.F) {
	f.Add([]byte(`target_key: "file_val"`), "env_val", "def_val")

	f.Fuzz(func(t *testing.T, data []byte, envVal string, defaultVal string) {
		if strings.ContainsAny(envVal, "=\n\r\x00") {
			return
		}

		const envKey = "VIPER_FUZZ_ENV_KEY"
		t.Setenv(envKey, envVal)

		v := New()
		v.SetConfigType("yaml")
		v.SetDefault("target_key", defaultVal)

		if err := v.BindEnv("target_key", envKey); err != nil {
			t.Fatal(err)
		}

		_ = v.ReadConfig(bytes.NewReader(data))

		// Verify priority: Environment variables should override config and defaults
		got := v.GetString("target_key")
		if envVal != "" {
			if got != envVal {
				t.Errorf("Priority error: Env should win. got %q, want %q", got, envVal)
			}
		}

		_ = v.GetInt("target_key")
		_ = v.GetBool("target_key")
		_ = v.GetStringSlice("target_key")
	})
}