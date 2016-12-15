// +build !go1.7

// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// This file is to return errors when attempting to use unsupported features.
// You can put its supported counterparts into go1.7Supported.go.

package viper

import (
	"errors"

	"github.com/spf13/afero"
)

// Marshal a map into Writer.
func marshalWriterHCL(f afero.File) error {
	return v.marshalWriterHCL(f)
}
func (v *Viper) marshalWriterHCL(f afero.File) error {
	return ConfigMarshalError{errors.New("HCL output unsupported before Go 1.7")}
}
