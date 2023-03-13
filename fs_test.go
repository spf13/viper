//go:build go1.16 && finder
// +build go1.16,finder

package viper

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinder(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"home/user/.config":            &fstest.MapFile{},
		"home/user/config.json":        &fstest.MapFile{},
		"home/user/config.yaml":        &fstest.MapFile{},
		"home/user/data.json":          &fstest.MapFile{},
		"etc/config/.config":           &fstest.MapFile{},
		"etc/config/a_random_file.txt": &fstest.MapFile{},
		"etc/config/config.json":       &fstest.MapFile{},
		"etc/config/config.yaml":       &fstest.MapFile{},
		"etc/config/config.xml":        &fstest.MapFile{},
	}

	testCases := []struct {
		name   string
		fsys   func() fs.FS
		finder finder
		result string
	}{
		{
			name: "find file",
			fsys: func() fs.FS { return fsys },
			finder: finder{
				paths:      []string{"etc/config"},
				fileNames:  []string{"config"},
				extensions: []string{"json"},
			},
			result: "etc/config/config.json",
		},
		{
			name: "file not found",
			fsys: func() fs.FS { return fsys },
			finder: finder{
				paths:      []string{"var/config"},
				fileNames:  []string{"config"},
				extensions: []string{"json"},
			},
			result: "",
		},
		{
			name:   "empty search params",
			fsys:   func() fs.FS { return fsys },
			finder: finder{},
			result: "",
		},
		{
			name: "precedence",
			fsys: func() fs.FS { return fsys },
			finder: finder{
				paths:      []string{"var/config", "home/user", "etc/config"},
				fileNames:  []string{"aconfig", "config"},
				extensions: []string{"zml", "xml", "json"},
			},
			result: "home/user/config.json",
		},
		{
			name: "without extension",
			fsys: func() fs.FS { return fsys },
			finder: finder{
				paths:      []string{"var/config", "home/user", "etc/config"},
				fileNames:  []string{".config"},
				extensions: []string{"zml", "xml", "json"},

				withoutExtension: true,
			},
			result: "home/user/.config",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			fsys := testCase.fsys()

			result, err := testCase.finder.Find(fsys)
			require.NoError(t, err)

			assert.Equal(t, testCase.result, result)
		})
	}
}
