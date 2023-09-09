//go:build go1.16 && finder
// +build go1.16,finder

package viper

import (
	"fmt"

	"github.com/sagikazarmark/go-finder"
	"github.com/spf13/afero"
)

// Search all configPaths for any config file.
// Returns the first path that exists (and is a config file).
func (v *Viper) findConfigFile() (string, error) {
	var names []string

	if v.configType != "" {
		names = finder.NameWithOptionalExtensions(v.configName, SupportedExts...)
	} else {
		names = finder.NameWithExtensions(v.configName, SupportedExts...)
	}

	finder := finder.Finder{
		Paths: v.configPaths,
		Names: names,
		Type:  finder.FileTypeFile,
	}

	results, err := finder.Find(afero.NewIOFS(v.fs))
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", ConfigFileNotFoundError{v.configName, fmt.Sprintf("%s", v.configPaths)}
	}

	return results[0], nil
}
