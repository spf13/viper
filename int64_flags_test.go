package viper_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/spf13/pflag"

	"github.com/spf13/viper"
)

func TestInt64PFlagsPreserveWidth(t *testing.T) {
	for _, value := range []int64{1 << 40, -(1 << 40), math.MaxInt64, math.MinInt64} {
		for _, changed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/changed=%t", value, changed), func(t *testing.T) {
				flags := pflag.NewFlagSet("limits", pflag.ContinueOnError)
				flags.Int64("max-bytes", value, "maximum bytes")
				if changed {
					if err := flags.Parse([]string{"--max-bytes=" + fmt.Sprint(value)}); err != nil {
						t.Fatal(err)
					}
				}
				v := viper.New()
				if err := v.BindPFlags(flags); err != nil {
					t.Fatal(err)
				}
				if got := v.GetInt64("max-bytes"); got != value {
					t.Fatalf("GetInt64 = %d, want %d", got, value)
				}
				var config struct {
					MaxBytes int64 `mapstructure:"max-bytes"`
				}
				if err := v.Unmarshal(&config); err != nil {
					t.Fatal(err)
				}
				if config.MaxBytes != value {
					t.Fatalf("Unmarshal = %d, want %d", config.MaxBytes, value)
				}
			})
		}
	}
}
