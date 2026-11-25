package viper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubAutomaticEnvWithCustomDelimiter(t *testing.T) {
	for _, delimiter := range []string{".", "::", "/"} {
		t.Run(delimiter, func(t *testing.T) {
			v := NewWithOptions(KeyDelimiter(delimiter))
			v.SetDefault(strings.Join([]string{"parent", "child", "value"}, delimiter), "default")
			v.SetEnvPrefix("APP")
			v.SetEnvKeyReplacer(strings.NewReplacer(delimiter, "_"))
			v.AutomaticEnv()
			t.Setenv("APP_PARENT_CHILD_VALUE", "environment")
			assert.Equal(t, "environment", v.GetString(strings.Join([]string{"parent", "child", "value"}, delimiter)))

			parent := v.Sub("parent")
			require.NotNil(t, parent)
			assert.Equal(t, "environment", parent.GetString("child"+delimiter+"value"))

			child := parent.Sub("child")
			require.NotNil(t, child)
			assert.Equal(t, "environment", child.GetString("value"))

			compound := v.Sub("parent" + delimiter + "child")
			require.NotNil(t, compound)
			assert.Equal(t, "environment", compound.GetString("value"))
		})
	}
}
