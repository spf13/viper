// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package viper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/kr/pretty"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v1"
)

// A set of paths to look for the config file in
var configPaths []string

// Name of file to look for inside the path
var configName string = "config"

// extensions Supported
var SupportedExts []string = []string{"json", "toml", "yaml"}
var configFile string
var configType string

var config map[string]interface{} = make(map[string]interface{})
var override map[string]interface{} = make(map[string]interface{})
var defaults map[string]interface{} = make(map[string]interface{})
var pflags map[string]*pflag.Flag = make(map[string]*pflag.Flag)
var aliases map[string]string = make(map[string]string)

// Explicitly define the path, name and extension of the config file
// Viper will use this and not check any of the config paths
func SetConfigFile(in string) {
	if in != "" {
		configFile = in
	}
}

func ConfigFileUsed() string {
	return configFile
}

// Add a path for viper to search for the config file in.
// Can be called multiple times to define multiple search paths.

func AddConfigPath(in string) {
	if in != "" {
		absin := absPathify(in)
		jww.INFO.Println("adding", absin, "to paths to search")
		if !stringInSlice(absin, configPaths) {
			configPaths = append(configPaths, absin)
		}
	}
}

func GetString(key string) string {
	return cast.ToString(Get(key))
}

func GetBool(key string) bool {
	return cast.ToBool(Get(key))
}

func GetInt(key string) int {
	return cast.ToInt(Get(key))
}

func GetFloat64(key string) float64 {
	return cast.ToFloat64(Get(key))
}

func GetTime(key string) time.Time {
	return cast.ToTime(Get(key))
}

func GetStringSlice(key string) []string {
	return cast.ToStringSlice(Get(key))
}

func GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(Get(key))
}

func GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(Get(key))
}

// Takes a single key and marshals it into a Struct
func MarshalKey(key string, rawVal interface{}) error {
	return mapstructure.Decode(Get(key), rawVal)
}

// Marshals the config into a Struct
func Marshal(rawVal interface{}) error {
	err := mapstructure.Decode(defaults, rawVal)
	if err != nil {
		return err
	}
	err = mapstructure.Decode(config, rawVal)
	if err != nil {
		return err 
	}
	err = mapstructure.Decode(override, rawVal)
	if err != nil {
		return err
	}

	insensativiseMaps()

	return nil
}

// Bind a specific key to a flag (as used by cobra)
//
//	 serverCmd.Flags().Int("port", 1138, "Port to run Application server on")
//	 viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))
//
func BindPFlag(key string, flag *pflag.Flag) (err error) {
	if flag == nil {
		return fmt.Errorf("flag for %q is nil", key)
	}
	pflags[key] = flag

	switch flag.Value.Type() {
	case "int", "int8", "int16", "int32", "int64":
		SetDefault(key, cast.ToInt(flag.Value.String()))
	case "bool":
		SetDefault(key, cast.ToBool(flag.Value.String()))
	default:
		SetDefault(key, flag.Value.String())
	}
	return nil
}

func find(key string) interface{} {
	var val interface{}
	var exists bool

	// if the requested key is an alias, then return the proper key
	key = realKey(key)

	// PFlag Override first
	flag, exists := pflags[key]
	if exists {
		if flag.Changed {
			jww.TRACE.Println(key, "found in override (via pflag):", val)
			return flag.Value.String()
		}
	}

	val, exists = override[key]
	if exists {
		jww.TRACE.Println(key, "found in override:", val)
		return val
	}

	val, exists = config[key]
	if exists {
		jww.TRACE.Println(key, "found in config:", val)
		return val
	}

	val, exists = defaults[key]
	if exists {
		jww.TRACE.Println(key, "found in defaults:", val)
		return val
	}

	return nil
}

// Get returns an interface..
// Must be typecast or used by something that will typecast
func Get(key string) interface{} {
	key = strings.ToLower(key)
	v := find(key)

	if v == nil {
		return nil
	}

	switch v.(type) {
	case bool:
		return cast.ToBool(v)
	case string:
		return cast.ToString(v)
	case int64, int32, int16, int8, int:
		return cast.ToInt(v)
	case float64, float32:
		return cast.ToFloat64(v)
	case time.Time:
		return cast.ToTime(v)
	case []string:
		return v
	}
	return v
}

func IsSet(key string) bool {
	t := Get(key)
	return t != nil
}

// Aliases provide another accessor for the same key.
// This enables one to change a name without breaking the application
func RegisterAlias(alias string, key string) {
	registerAlias(alias, strings.ToLower(key))
}

func registerAlias(alias string, key string) {
	alias = strings.ToLower(alias)
	if alias != key && alias != realKey(key) {
		_, exists := aliases[alias]

		if !exists {
			// if we alias something that exists in one of the maps to another
			// name, we'll never be able to get that value using the original
			// name, so move the config value to the new realkey.
			if val, ok := config[alias]; ok {
				delete(config, alias)
				config[key] = val
			}
			if val, ok := defaults[alias]; ok {
				delete(defaults, alias)
				defaults[key] = val
			}
			if val, ok := override[alias]; ok {
				delete(override, alias)
				override[key] = val
			}
			aliases[alias] = key
		}
	} else {
		jww.WARN.Println("Creating circular reference alias", alias, key, realKey(key))
	}
}

func realKey(key string) string {
	newkey, exists := aliases[key]
	if exists {
		jww.DEBUG.Println("Alias", key, "to", newkey)
		return realKey(newkey)
	} else {
		return key
	}
}

func InConfig(key string) bool {
	// if the requested key is an alias, then return the proper key
	key = realKey(key)

	_, exists := config[key]
	return exists
}

// Set the default value for this key.
// Default only used when no value is provided by the user via flag, config or ENV.
func SetDefault(key string, value interface{}) {
	// If alias passed in, then set the proper default
	key = realKey(strings.ToLower(key))
	defaults[key] = value
}

// The user provided value (via flag)
// Will be used instead of values obtained via config file, ENV or default
func Set(key string, value interface{}) {
	// If alias passed in, then set the proper override
	key = realKey(strings.ToLower(key))
	override[key] = value
}

type UnsupportedConfigError string

func (str UnsupportedConfigError) Error() string {
	return fmt.Sprintf("Unsupported Config Type %q", string(str))
}

// Viper will discover and load the configuration file from disk
// searching in one of the defined paths.
func ReadInConfig() error {
	jww.INFO.Println("Attempting to read in config file")
	if !stringInSlice(getConfigType(), SupportedExts) {
		return UnsupportedConfigError(getConfigType())
	}

	file, err := ioutil.ReadFile(getConfigFile())
	if err != nil {
		return err
	}

	MarshallReader(bytes.NewReader(file))
	return nil
}

func MarshallReader(in io.Reader) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	switch getConfigType() {
	case "yaml", "yml":
		if err := yaml.Unmarshal(buf.Bytes(), &config); err != nil {
			jww.ERROR.Fatalf("Error parsing config: %s", err)
		}

	case "json":
		if err := json.Unmarshal(buf.Bytes(), &config); err != nil {
			jww.ERROR.Fatalf("Error parsing config: %s", err)
		}

	case "toml":
		if _, err := toml.Decode(buf.String(), &config); err != nil {
			jww.ERROR.Fatalf("Error parsing config: %s", err)
		}
	}

	insensativiseMap(config)
}

func insensativiseMaps() {
	insensativiseMap(config)
	insensativiseMap(defaults)
	insensativiseMap(override)
}

func insensativiseMap(m map[string]interface{}) {
	for key, val := range m {
		lower := strings.ToLower(key)
		if key != lower {
			delete(m, key)
			m[lower] = val
		}
	}
}


// Name for the config file.
// Does not include extension.
func SetConfigName(in string) {
	if in != "" {
		configName = in
	}
}

func SetConfigType(in string) {
	if in != "" {
		configType = in
	}
}

func getConfigType() string {
	if configType != "" {
		return configType
	}

	cf := getConfigFile()
	ext := path.Ext(cf)

	if len(ext) > 1 {
		return ext[1:]
	} else {
		return ""
	}
}

func getConfigFile() string {
	// if explicitly set, then use it
	if configFile != "" {
		return configFile
	}

	cf, err := findConfigFile()
	if err != nil {
		return ""
	}

	configFile = cf
	return getConfigFile()
}

func searchInPath(in string) (filename string) {
	jww.DEBUG.Println("Searching for config in ", in)
	for _, ext := range SupportedExts {

		jww.DEBUG.Println("Checking for", path.Join(in, configName+"."+ext))
		if b, _ := exists(path.Join(in, configName+"."+ext)); b {
			jww.DEBUG.Println("Found: ", path.Join(in, configName+"."+ext))
			return path.Join(in, configName+"."+ext)
		}
	}

	return ""
}

func findConfigFile() (string, error) {
	jww.INFO.Println("Searching for config in ", configPaths)

	for _, cp := range configPaths {
		file := searchInPath(cp)
		if file != "" {
			return file, nil
		}
	}
	cwd, _ := findCWD()

	file := searchInPath(cwd)
	if file != "" {
		return file, nil
	}

	return "", fmt.Errorf("config file not found in: %s", configPaths)
}

func findCWD() (string, error) {
	serverFile, err := filepath.Abs(os.Args[0])

	if err != nil {
		return "", fmt.Errorf("Can't get absolute path for executable: %v", err)
	}

	path := filepath.Dir(serverFile)
	realFile, err := filepath.EvalSymlinks(serverFile)

	if err != nil {
		if _, err = os.Stat(serverFile + ".exe"); err == nil {
			realFile = filepath.Clean(serverFile + ".exe")
		}
	}

	if err == nil && realFile != serverFile {
		path = filepath.Dir(realFile)
	}

	return path, nil
}

// Check if File / Directory Exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func absPathify(inPath string) string {
	jww.INFO.Println("Trying to resolve absolute path to", inPath)

	if strings.HasPrefix(inPath, "$HOME") {
		inPath = userHomeDir() + inPath[5:]
	}

	if strings.HasPrefix(inPath, "$") {
		end := strings.Index(inPath, string(os.PathSeparator))
		inPath = os.Getenv(inPath[1:end]) + inPath[end:]
	}

	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}

	p, err := filepath.Abs(inPath)
	if err == nil {
		return filepath.Clean(p)
	} else {
		jww.ERROR.Println("Couldn't discover absolute path")
		jww.ERROR.Println(err)
	}
	return ""
}

func Debug() {
	fmt.Println("Config:")
	pretty.Println(config)
	fmt.Println("Defaults:")
	pretty.Println(defaults)
	fmt.Println("Override:")
	pretty.Println(override)
	fmt.Println("Aliases:")
	pretty.Println(aliases)
}

func Reset() {
	configPaths = nil
	configName = "config"

	// extensions Supported
	SupportedExts = []string{"json", "toml", "yaml", "yml"}
	configFile = ""
	configType = ""

	config = make(map[string]interface{})
	override = make(map[string]interface{})
	defaults = make(map[string]interface{})
	aliases = make(map[string]string)
}
