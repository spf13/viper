// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package viper

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestGetConfigView(t *testing.T) {
	clothingPantsSize := "small"
	os.Setenv("SPF_CLOTHING_PANTS_SIZE", clothingPantsSize)

	v := New()
	v.SetConfigType("yaml")
	v.ReadConfig(bytes.NewBuffer(yamlExample))

	v.SetEnvPrefix("spf")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Sub fail to get env override
	subv := v.Sub("clothing")
	assert.NotEqual(t, v.Get("clothing.pants.size"), subv.Get("pants.size"))

	subConfig := v.GetConfigView("clothing")
	assert.Equal(t, v.Get("clothing.pants.size"), subConfig.Get("pants.size"))

	sub2Config := subConfig.GetConfigView("pants")
	assert.Equal(t, v.Get("clothing.pants.size"), sub2Config.Get("size"))
}
