package viper

import (
	"fmt"
)

/* ConfigFileNotFoundError types */

// For matching on any ConfigFileNotFoundFrom*Error
type ConfigFileNotFoundError interface {
	Error() string
	Name() string
	Locations() string
}

// ConfigFileNotFoundError denotes failing to find configuration file.
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

// ConfigFileNotFoundError denotes failing to find configuration file.
type ConfigFileNotFoundFromReadError struct {
	name string
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
