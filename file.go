//go:build !finder

package viper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// Search all configPaths for any config file.
// Returns the first path that exists (and is a config file).
func (v *Viper) findConfigFile() (string, error) {
	v.logger.Info("searching for config in paths", "paths", v.configPaths)

	for _, cp := range v.configPaths {
		file := v.searchInPath(cp)
		if file != "" {
			return file, nil
		}
	}
	return "", ConfigFileNotFoundError{v.configName, fmt.Sprintf("%s", v.configPaths)}
}

func (v *Viper) searchInPath(in string) (filename string) {
	v.logger.Debug("searching for config in path", "path", in)
	for _, ext := range SupportedExts {
		v.logger.Debug("checking if file exists", "file", filepath.Join(in, v.configName+"."+ext))
		if b, _ := exists(v.fs, filepath.Join(in, v.configName+"."+ext)); b {
			// record the first found
			if len(filename) < 1 {
				filename = filepath.Join(in, v.configName+"."+ext)
			}
			// if specific configType are same with the current extension type, then return current
			if v.configType == ext {
				filename = filepath.Join(in, v.configName+"."+ext)
				break
			}
		}
	}

	// return only if file exists
	if len(filename) > 0 {
		v.logger.Debug("found file", "file", filename)
		return filename
	}

	if v.configType != "" {
		if b, _ := exists(v.fs, filepath.Join(in, v.configName)); b {
			return filepath.Join(in, v.configName)
		}
	}

	return ""
}

// exists checks if file exists.
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
