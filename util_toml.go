// +build toml

// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Viper is a application configuration system.
// It believes that applications can be configured a variety of ways
// via flags, ENVIRONMENT variables, configuration files retrieved
// from the file system, or a remote key/value store.

package viper

import (
	"github.com/BurntSushi/toml"
)

func unmarshalTOML(data string, v interface{}) error {
	if _, err := toml.Decode(data, v); err != nil {
		return err
	}
	return nil
}
