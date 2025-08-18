package viper

import (
	"fmt"
)

/* File look-up errors */

// FileLookupError is returned when Viper cannot resolve a configuration file.
//
// This is meant to be a common interface for all file look-up errors, occurring either because a
// file does not exist or because it cannot find any file matching finder criteria.
type FileLookupError interface {
	error

	fileLookup()
}

// ConfigFileNotFoundError denotes failing to find a configuration file from a search.
//
// Deprecated: This is wrapped by FileNotFoundFromSearchError, which should be used instead.
type ConfigFileNotFoundError struct {
	locations []string
	name      string
}

// Error returns the formatted error.
func (e ConfigFileNotFoundError) Error() string {
	message := fmt.Sprintf("File %q Not Found", e.name)
	if len(e.locations) != 0 {
		message += fmt.Sprintf(" in %v", e.locations)
	}

	return message
}

// Unwraps to FileNotFoundFromSearchError.
func (e ConfigFileNotFoundError) Unwrap() error {
	return FileNotFoundFromSearchError{err: e, locations: e.locations, name: e.name}
}

// FileNotFoundFromSearchError denotes failing to find a configuration file from a search.
// Wraps ConfigFileNotFoundError.
type FileNotFoundFromSearchError struct {
	err       ConfigFileNotFoundError
	locations []string
	name      string
}

func (e FileNotFoundFromSearchError) fileLookup() {
	return
}

// Error returns the formatted error.
func (e FileNotFoundFromSearchError) Error() string {
	return e.err.Error()
}

// FileNotFoundError denotes failing to find a specific configuration file.
type FileNotFoundError struct {
	err  error
	path string
}

func (e FileNotFoundError) fileLookup() {
	return
}

// Error returns the formatted error.
func (e FileNotFoundError) Error() string {
	return fmt.Sprintf("file not found: %s", e.path)
}

/* Other error types */

// ConfigFileAlreadyExistsError denotes failure to write new configuration file.
type ConfigFileAlreadyExistsError string

// Error returns the formatted error when configuration already exists.
func (e ConfigFileAlreadyExistsError) Error() string {
	return fmt.Sprintf("Config File %q Already Exists", string(e))
}

// ConfigMarshalError happens when failing to marshal the configuration.
type ConfigMarshalError struct {
	err error
}

// Error returns the formatted configuration error.
func (e ConfigMarshalError) Error() string {
	return fmt.Sprintf("While marshaling config: %s", e.err.Error())
}

// UnsupportedConfigError denotes encountering an unsupported
// configuration filetype.
type UnsupportedConfigError string

// Error returns the formatted configuration error.
func (str UnsupportedConfigError) Error() string {
	return fmt.Sprintf("Unsupported Config Type %q", string(str))
}
