package viper

import (
	"fmt"
)

/* File look-up errors */

// ConfigFileNotFoundError denotes failing to find a configuration file from a search.
//
// Deprecated: This is wrapped by FileNotFoundFromSearchError, which should be used instead.
type ConfigFileNotFoundError struct {
	name      string
	locations []string
}

// Error returns the formatted error.
func (fnfe ConfigFileNotFoundError) Error() string {
	message := fmt.Sprintf("File %q Not Found", fnfe.name)
	if len(fnfe.locations) != 0 {
		message += fmt.Sprintf(" in %v", fnfe.locations)
	}

	return message
}

// Unwraps to FileNotFoundFromSearchError.
func (fnfe ConfigFileNotFoundError) Unwrap() error {
	return FileNotFoundFromSearchError{err: fnfe}
}

// FileNotFoundFromSearchError denotes failing to find a configuration file from a search.
// Wraps ConfigFileNotFoundError.
type FileNotFoundFromSearchError struct {
	err ConfigFileNotFoundError
}

// Error returns the formatted error.
func (fnfe FileNotFoundFromSearchError) Error() string {
	return fnfe.err.Error()
}

// FileNotFoundError denotes failing to find a specific configuration file.
type FileNotFoundError struct {
	path string
}

// Error returns the formatted error.
func (fnfe FileNotFoundError) Error() string {
	return fmt.Sprintf("File %q Not Found", fnfe.path)
}

/* Other error types */

// ConfigFileAlreadyExistsError denotes failure to write new configuration file.
type ConfigFileAlreadyExistsError string

// Error returns the formatted error when configuration already exists.
func (faee ConfigFileAlreadyExistsError) Error() string {
	return fmt.Sprintf("Config File %q Already Exists", string(faee))
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
