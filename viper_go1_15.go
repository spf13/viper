//go:build !go1.16 || !finder
// +build !go1.16 !finder

package viper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	jww "github.com/spf13/jwalterweatherman"
)

// Search all configPaths for any config file.
// Returns the first path that exists (and is a config file).
func (v *Viper) findConfigFile() (string, error) {
	jww.INFO.Println("Searching for config in ", v.configPaths)

	for _, cp := range v.configPaths {
		file := v.searchInPath(cp)
		if file != "" {
			return file, nil
		}
	}
	return "", ConfigFileNotFoundError{v.configName, fmt.Sprintf("%s", v.configPaths)}
}

func (v *Viper) searchInPath(in string) (filename string) {
	jww.DEBUG.Println("Searching for config in ", in)
	for _, ext := range SupportedExts {
		jww.DEBUG.Println("Checking for", filepath.Join(in, v.configName+"."+ext))
		if b, _ := exists(v.fs, filepath.Join(in, v.configName+"."+ext)); b {
			jww.DEBUG.Println("Found: ", filepath.Join(in, v.configName+"."+ext))
			return filepath.Join(in, v.configName+"."+ext)
		}
	}

	if v.configType != "" {
		if b, _ := exists(v.fs, filepath.Join(in, v.configName)); b {
			return filepath.Join(in, v.configName)
		}
	}

	return ""
}

// Check if file Exists
func exists(fs afero.Fs, path string) (bool, error) {
	stat, err := fs.Stat(path)
	if err == nil {
		return !stat.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
