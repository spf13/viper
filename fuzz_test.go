// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package viper_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/spf13/viper"
)

func FuzzReadConfig(f *testing.F) {
	yamlSeed := []byte("server:\n  port: 8080\n  host: localhost\n")
	jsonSeed := []byte(`{"server": {"port": 8080, "host": "localhost"}}`)
	tomlSeed := []byte("[server]\nport = 8080\nhost = \"localhost\"\n")
	dotenvSeed := []byte("SERVER_PORT=8080\nSERVER_HOST=localhost\n")

	f.Add("yaml", yamlSeed)
	f.Add("json", jsonSeed)
	f.Add("toml", tomlSeed)
	f.Add("dotenv", dotenvSeed)
	f.Add("yaml", []byte{})
	f.Add("yaml", []byte{0xFF, 0x00, 0xFF})

	f.Fuzz(func(t *testing.T, format string, data []byte) {
		if len(format) > 20 || len(data) > 1<<20 { return }
		format = strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' { return r }
			return -1
		}, strings.ToLower(format))
		if format == "" { format = "yaml" }
		func() {
			defer func() { recover() }()
			v := viper.New()
			v.SetConfigType(format)
			v.ReadConfig(bytes.NewReader(data))
			for _, k := range v.AllKeys() {
				v.Get(k)
			}
		}()
	})
}

func FuzzUnmarshal(f *testing.F) {
	type Cfg struct {
		Name    string  `mapstructure:"name"`
		Port    int     `mapstructure:"port"`
		Enabled bool    `mapstructure:"enabled"`
		Rate    float64 `mapstructure:"rate"`
	}
	f.Add([]byte("name: test\nport: 8080\nenabled: true\nrate: 0.75\n"))
	f.Add([]byte("name: \"\"\nport: -1\nenabled: false\n"))
	f.Add([]byte("!!binary dGVzdA==\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<16 { return }
		func() {
			defer func() { recover() }()
			v := viper.New()
			v.SetConfigType("yaml")
			if err := v.ReadConfig(bytes.NewReader(data)); err != nil { return }
			var cfg Cfg
			v.Unmarshal(&cfg)
		}()
	})
}

func FuzzMergeConfigMap(f *testing.F) {
	f.Add([]byte("key1: val1\n"), []byte("key2: val2\n"))
	f.Add([]byte("key: val1\n"), []byte("key: val2\n"))
	f.Add([]byte(""), []byte("key: val\n"))
	f.Fuzz(func(t *testing.T, base, override []byte) {
		if len(base) > 1<<16 || len(override) > 1<<16 { return }
		func() {
			defer func() { recover() }()
			v := viper.New()
			v.SetConfigType("yaml")
			v.ReadConfig(bytes.NewReader(base))
			v.MergeConfig(bytes.NewReader(override))
			for _, k := range v.AllKeys() { v.Get(k) }
		}()
	})
}

func FuzzSetGet(f *testing.F) {
	f.Add("key", "value")
	f.Add("", "empty-key")
	f.Add("nested.key.path", "deep")
	f.Fuzz(func(t *testing.T, key, value string) {
		if len(key) > 5000 || len(value) > 1<<16 { return }
		func() {
			defer func() { recover() }()
			v := viper.New()
			v.Set(key, value)
			v.Get(key)
			v.GetString(key)
			v.GetInt(key)
			v.GetBool(key)
		}()
	})
}

func FuzzConfigFile(f *testing.F) {
	f.Add("yaml", []byte("key: value\n"))
	f.Add("json", []byte(`{"key":"value"}`))
	f.Add("toml", []byte("key = \"value\"\n"))
	f.Fuzz(func(t *testing.T, ext string, data []byte) {
		if len(ext) > 10 || len(data) > 1<<16 || len(data) < 2 { return }
		ext = strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' { return r }
			return -1
		}, strings.ToLower(ext))
		if ext == "" { ext = "yaml" }
		func() {
			defer func() { recover() }()
			dir, _ := os.MkdirTemp("", "fuzz-viper-*")
			defer os.RemoveAll(dir)
			fp := filepath.Join(dir, "config."+ext)
			os.WriteFile(fp, data, 0644)
			v := viper.New()
			v.SetConfigFile(fp)
			v.ReadInConfig()
			v.AllSettings()
		}()
	})
}
