// +build go1.7

// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// This file is to house features only supported since Go 1.7.
// You can put its counterpart in go1.7Unsupported.go.

package viper

import (
	"encoding/json"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/spf13/afero"
)

// Marshal a map into Writer.
func marshalWriterHCL(f afero.File) error {
	return v.marshalWriterHCL(f)
}
func (v *Viper) marshalWriterHCL(f afero.File) error {
	c := v.config
	b, err := json.Marshal(c)
	ast, err := hcl.Parse(string(b))
	if err != nil {
		return ConfigMarshalError{err}
	}
	err = printer.Fprint(f, ast.Node)
	if err != nil {
		return ConfigMarshalError{err}
	}
	return nil
}
