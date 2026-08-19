package viper

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type finderStub struct {
	results []string
}

func (f finderStub) Find(_ afero.Fs) ([]string, error) {
	return f.results, nil
}

type errorFinderStub struct {
	err error
}

func (f errorFinderStub) Find(_ afero.Fs) ([]string, error) {
	return nil, f.err
}

func TestFinders(t *testing.T) {
	finder := Finders(
		finderStub{
			results: []string{
				"/home/user/.viper.yaml",
			},
		},
		finderStub{
			results: []string{
				"/etc/viper/config.yaml",
			},
		},
	)

	results, err := finder.Find(afero.NewMemMapFs())
	require.NoError(t, err)

	expected := []string{
		"/home/user/.viper.yaml",
		"/etc/viper/config.yaml",
	}

	assert.Equal(t, expected, results)
}

func TestFinders_WithNilAndErrors(t *testing.T) {
	errDummy := errors.New("find error")
	finder := Finders(
		nil,
		finderStub{
			results: []string{
				"/home/user/.viper.yaml",
			},
		},
		errorFinderStub{
			err: errDummy,
		},
		finderStub{
			results: []string{
				"/etc/viper/config.yaml",
			},
		},
	)

	results, err := finder.Find(afero.NewMemMapFs())
	require.Error(t, err)
	assert.ErrorIs(t, err, errDummy)

	expected := []string{
		"/home/user/.viper.yaml",
		"/etc/viper/config.yaml",
	}
	assert.Equal(t, expected, results)
}

func TestWithFinderOption(t *testing.T) {
	v := New()
	optNil := WithFinder(nil)
	optNil.apply(v)
	assert.Nil(t, v.finder)

	stub := finderStub{results: []string{"/test.yaml"}}
	opt := WithFinder(stub)
	opt.apply(v)
	assert.Equal(t, stub, v.finder)
}
