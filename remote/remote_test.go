package remote

import (
	"testing"

	"github.com/spf13/viper"
)

func Test_init(t *testing.T) {
	if viper.RemoteConfig == nil {
		t.Fatal("viper.RemoteConfig() is nil, want non nil")
	}
}
