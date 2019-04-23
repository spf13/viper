package viper

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestBindFlagValueSet(t *testing.T) {
	v := New()

	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)

	var testValues = map[string]*string{
		"host":     nil,
		"port":     nil,
		"endpoint": nil,
	}

	var mutatedTestValues = map[string]string{
		"host":     "localhost",
		"port":     "6060",
		"endpoint": "/public",
	}

	for name := range testValues {
		testValues[name] = flagSet.String(name, "", "test")
	}

	flagValueSet := pflagValueSet{flagSet}

	err := v.BindFlagValues(flagValueSet)
	if err != nil {
		t.Fatalf("error binding flag set, %v", err)
	}

	flagSet.VisitAll(func(flag *pflag.Flag) {
		flag.Value.Set(mutatedTestValues[flag.Name])
		flag.Changed = true
	})

	for name, expected := range mutatedTestValues {
		assert.Equal(t, v.Get(name), expected)
	}
}

func TestBindFlagValue(t *testing.T) {
	v := New()

	var testString = "testing"
	var testValue = newStringValue(testString, &testString)

	flag := &pflag.Flag{
		Name:    "testflag",
		Value:   testValue,
		Changed: false,
	}

	flagValue := pflagValue{flag}
	v.BindFlagValue("testvalue", flagValue)

	assert.Equal(t, testString, v.Get("testvalue"))

	flag.Value.Set("testing_mutate")
	flag.Changed = true //hack for pflag usage

	assert.Equal(t, "testing_mutate", v.Get("testvalue"))
}
