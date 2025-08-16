package viper

import (
	"fmt"
)


// FileLookupError is returned when Viper cannot resolve a configuration file,
// either because a file does not exists or because it cannot find any file matching the criteria.
type FileLookupError interface {
	error
	
	fileLookup()
}

// ConfigFileNotFoundFromFinderError denotes failing to find a configuration file.
type ConfigFileNotFoundFromFinderError struct {
	name, locations string
}

// Error returns the formatted configuration error.
func (fnfe ConfigFileNotFoundFromFinderError) Error() string {
	message := fmt.Sprintf("Config file %q Not Found", fnfe.name)
	if fnfe.locations != "" {
		message += fmt.Sprintf(" in %q", fnfe.locations)
	}

	return message
}

func (fnfe ConfigFileNotFoundFromFinderError) Name() string {
	return fnfe.name
}

func (fnfe ConfigFileNotFoundFromFinderError) Locations() string {
	return fnfe.locations
}

// ConfigFileNotFoundFromReadError denotes failing to find a specific configuration file.
type FileNotFoundError struct {
	path string
}

// Error returns the formatted configuration error.
func (fnfe ConfigFileNotFoundFromReadError) Error() string {
	return fmt.Sprintf("Config file %q Not Found", fnfe.name)
}

func (fnfe ConfigFileNotFoundFromReadError) Name() string {
	return fnfe.name
}

func (fnfe ConfigFileNotFoundFromReadError) Locations() string {
	return ""
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
