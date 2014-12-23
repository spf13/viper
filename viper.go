// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Viper is a application configuration system.
// It believes that applications can be configured a variety of ways
// via flags, ENVIRONMENT variables, configuration files retrieved
// from the file system, or a remote key/value store.

// Each item takes precedence over the item below it:

// flag
// env
// config
// key/value store
// default

package viper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/pflag"
	crypt "github.com/xordataexchange/crypt/config"
)

var v *viper

func init() {
	v = New()
}

type UnsupportedConfigError string

func (str UnsupportedConfigError) Error() string {
	return fmt.Sprintf("Unsupported Config Type %q", string(str))
}

type UnsupportedRemoteProviderError string

func (str UnsupportedRemoteProviderError) Error() string {
	return fmt.Sprintf("Unsupported Remote Provider Type %q", string(str))
}

type RemoteConfigError string

func (rce RemoteConfigError) Error() string {
	return fmt.Sprintf("Remote Configurations Error: %s", string(rce))
}

// A viper is a unexported struct. Use New() to create a new instance of viper
// or use the functions for a "global instance"
type viper struct {
	// A set of paths to look for the config file in
	configPaths []string

	// A set of remote providers to search for the configuration
	remoteProviders []*remoteProvider

	// Name of file to look for inside the path
	configName string
	configFile string
	configType string
	envPrefix  string

	automaticEnvApplied bool

	config   map[string]interface{}
	override map[string]interface{}
	defaults map[string]interface{}
	kvstore  map[string]interface{}
	pflags   map[string]*pflag.Flag
	env      map[string]string
	aliases  map[string]string
}

// The prescribed way to create a new Viper
func New() *viper {
	v := new(viper)
	v.configName = "config"
	v.config = make(map[string]interface{})
	v.override = make(map[string]interface{})
	v.defaults = make(map[string]interface{})
	v.kvstore = make(map[string]interface{})
	v.pflags = make(map[string]*pflag.Flag)
	v.env = make(map[string]string)
	v.aliases = make(map[string]string)

	return v
}

// remoteProvider stores the configuration necessary
// to connect to a remote key/value store.
// Optional secretKeyring to unencrypt encrypted values
// can be provided.
type remoteProvider struct {
	provider      string
	endpoint      string
	path          string
	secretKeyring string
}

// universally supported extensions
var SupportedExts []string = []string{"json", "toml", "yaml", "yml"}

// universally supported remote providers
var SupportedRemoteProviders []string = []string{"etcd", "consul"}

// Explicitly define the path, name and extension of the config file
// Viper will use this and not check any of the config paths
func SetConfigFile(in string) { v.SetConfigFile(in) }
func (v *viper) SetConfigFile(in string) {
	if in != "" {
		v.configFile = in
	}
}

// Define a prefix that ENVIRONMENT variables will use.
func SetEnvPrefix(in string) { v.SetEnvPrefix(in) }
func (v *viper) SetEnvPrefix(in string) {
	if in != "" {
		v.envPrefix = in
	}
}

func (v *viper) mergeWithEnvPrefix(in string) string {
	if v.envPrefix != "" {
		return strings.ToUpper(v.envPrefix + "_" + in)
	}

	return strings.ToUpper(in)
}

// Return the config file used
func ConfigFileUsed() string            { return v.ConfigFileUsed() }
func (v *viper) ConfigFileUsed() string { return v.configFile }

// Add a path for viper to search for the config file in.
// Can be called multiple times to define multiple search paths.
func AddConfigPath(in string) { v.AddConfigPath(in) }
func (v *viper) AddConfigPath(in string) {
	if in != "" {
		absin := absPathify(in)
		jww.INFO.Println("adding", absin, "to paths to search")
		if !stringInSlice(absin, v.configPaths) {
			v.configPaths = append(v.configPaths, absin)
		}
	}
}

// AddRemoteProvider adds a remote configuration source.
// Remote Providers are searched in the order they are added.
// provider is a string value, "etcd" or "consul" are currently supported.
// endpoint is the url.  etcd requires http://ip:port  consul requires ip:port
// path is the path in the k/v store to retrieve configuration
// To retrieve a config file called myapp.json from /configs/myapp.json
// you should set path to /configs and set config name (SetConfigName()) to
// "myapp"
func AddRemoteProvider(provider, endpoint, path string) error {
	return v.AddRemoteProvider(provider, endpoint, path)
}
func (v *viper) AddRemoteProvider(provider, endpoint, path string) error {
	if !stringInSlice(provider, SupportedRemoteProviders) {
		return UnsupportedRemoteProviderError(provider)
	}
	if provider != "" && endpoint != "" {
		jww.INFO.Printf("adding %s:%s to remote provider list", provider, endpoint)
		rp := &remoteProvider{
			endpoint: endpoint,
			provider: provider,
			path:     path,
		}
		if !v.providerPathExists(rp) {
			v.remoteProviders = append(v.remoteProviders, rp)
		}
	}
	return nil
}

// AddSecureRemoteProvider adds a remote configuration source.
// Secure Remote Providers are searched in the order they are added.
// provider is a string value, "etcd" or "consul" are currently supported.
// endpoint is the url.  etcd requires http://ip:port  consul requires ip:port
// secretkeyring is the filepath to your openpgp secret keyring.  e.g. /etc/secrets/myring.gpg
// path is the path in the k/v store to retrieve configuration
// To retrieve a config file called myapp.json from /configs/myapp.json
// you should set path to /configs and set config name (SetConfigName()) to
// "myapp"
// Secure Remote Providers are implemented with github.com/xordataexchange/crypt
func AddSecureRemoteProvider(provider, endpoint, path, secretkeyring string) error {
	return v.AddSecureRemoteProvider(provider, endpoint, path, secretkeyring)
}

func (v *viper) AddSecureRemoteProvider(provider, endpoint, path, secretkeyring string) error {
	if !stringInSlice(provider, SupportedRemoteProviders) {
		return UnsupportedRemoteProviderError(provider)
	}
	if provider != "" && endpoint != "" {
		jww.INFO.Printf("adding %s:%s to remote provider list", provider, endpoint)
		rp := &remoteProvider{
			endpoint: endpoint,
			provider: provider,
			path:     path,
		}
		if !v.providerPathExists(rp) {
			v.remoteProviders = append(v.remoteProviders, rp)
		}
	}
	return nil
}

func (v *viper) providerPathExists(p *remoteProvider) bool {
	for _, y := range v.remoteProviders {
		if reflect.DeepEqual(y, p) {
			return true
		}
	}
	return false
}

// Viper is essentially repository for configurations
// Get can retrieve any value given the key to use
// Get has the behavior of returning the value associated with the first
// place from where it is set. Viper will check in the following order:
// flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) interface{} { return v.Get(key) }
func (v *viper) Get(key string) interface{} {
	key = strings.ToLower(key)
	val := v.find(key)

	if val == nil {
		return nil
	}

	switch val.(type) {
	case bool:
		return cast.ToBool(val)
	case string:
		return cast.ToString(val)
	case int64, int32, int16, int8, int:
		return cast.ToInt(val)
	case float64, float32:
		return cast.ToFloat64(val)
	case time.Time:
		return cast.ToTime(val)
	case []string:
		return val
	}
	return val
}

func GetString(key string) string { return v.GetString(key) }
func (v *viper) GetString(key string) string {
	return cast.ToString(v.Get(key))
}

func GetBool(key string) bool { return v.GetBool(key) }
func (v *viper) GetBool(key string) bool {
	return cast.ToBool(v.Get(key))
}

func GetInt(key string) int { return v.GetInt(key) }
func (v *viper) GetInt(key string) int {
	return cast.ToInt(v.Get(key))
}

func GetFloat64(key string) float64 { return v.GetFloat64(key) }
func (v *viper) GetFloat64(key string) float64 {
	return cast.ToFloat64(v.Get(key))
}

func GetTime(key string) time.Time { return v.GetTime(key) }
func (v *viper) GetTime(key string) time.Time {
	return cast.ToTime(v.Get(key))
}

func GetStringSlice(key string) []string { return v.GetStringSlice(key) }
func (v *viper) GetStringSlice(key string) []string {
	return cast.ToStringSlice(v.Get(key))
}

func GetStringMap(key string) map[string]interface{} { return v.GetStringMap(key) }
func (v *viper) GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(v.Get(key))
}

func GetStringMapString(key string) map[string]string { return v.GetStringMapString(key) }
func (v *viper) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(v.Get(key))
}

// Takes a single key and marshals it into a Struct
func MarshalKey(key string, rawVal interface{}) error { return v.MarshalKey(key, rawVal) }
func (v *viper) MarshalKey(key string, rawVal interface{}) error {
	return mapstructure.Decode(v.Get(key), rawVal)
}

// Marshals the config into a Struct
func Marshal(rawVal interface{}) error { return v.Marshal(rawVal) }
func (v *viper) Marshal(rawVal interface{}) error {
	err := mapstructure.Decode(v.defaults, rawVal)
	if err != nil {
		return err
	}
	err = mapstructure.Decode(v.config, rawVal)
	if err != nil {
		return err
	}
	err = mapstructure.Decode(v.override, rawVal)
	if err != nil {
		return err
	}
	err = mapstructure.Decode(v.kvstore, rawVal)
	if err != nil {
		return err
	}

	v.insensativiseMaps()

	return nil
}

// Bind a specific key to a flag (as used by cobra)
//
//	 serverCmd.Flags().Int("port", 1138, "Port to run Application server on")
//	 viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))
//
func BindPFlag(key string, flag *pflag.Flag) (err error) { return v.BindPFlag(key, flag) }
func (v *viper) BindPFlag(key string, flag *pflag.Flag) (err error) {
	if flag == nil {
		return fmt.Errorf("flag for %q is nil", key)
	}
	v.pflags[strings.ToLower(key)] = flag

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

// Binds a viper key to a ENV variable
// ENV variables are case sensitive
// If only a key is provided, it will use the env key matching the key, uppercased.
// EnvPrefix will be used when set when env name is not provided.
func BindEnv(input ...string) (err error) { return v.BindEnv(input...) }
func (v *viper) BindEnv(input ...string) (err error) {
	var key, envkey string
	if len(input) == 0 {
		return fmt.Errorf("BindEnv missing key to bind to")
	}

	key = strings.ToLower(input[0])

	if len(input) == 1 {
		envkey = v.mergeWithEnvPrefix(key)
	} else {
		envkey = input[1]
	}

	v.env[key] = envkey

	return nil
}

// Given a key, find the value
// Viper will check in the following order:
// flag, env, config file, key/value store, default
// Viper will check to see if an alias exists first
func (v *viper) find(key string) interface{} {
	var val interface{}
	var exists bool

	// if the requested key is an alias, then return the proper key
	key = v.realKey(key)

	// PFlag Override first
	flag, exists := v.pflags[key]
	if exists {
		if flag.Changed {
			jww.TRACE.Println(key, "found in override (via pflag):", val)
			return flag.Value.String()
		}
	}

	val, exists = v.override[key]
	if exists {
		jww.TRACE.Println(key, "found in override:", val)
		return val
	}

	if v.automaticEnvApplied {
		// even if it hasn't been registered, if automaticEnv is used,
		// check any Get request
		if val = os.Getenv(v.mergeWithEnvPrefix(key)); val != "" {
			jww.TRACE.Println(key, "found in environment with val:", val)
			return val
		}
	}

	envkey, exists := v.env[key]
	if exists {
		jww.TRACE.Println(key, "registered as env var", envkey)
		if val = os.Getenv(envkey); val != "" {
			jww.TRACE.Println(envkey, "found in environment with val:", val)
			return val
		} else {
			jww.TRACE.Println(envkey, "env value unset:")
		}
	}

	val, exists = v.config[key]
	if exists {
		jww.TRACE.Println(key, "found in config:", val)
		return val
	}

	val, exists = v.kvstore[key]
	if exists {
		jww.TRACE.Println(key, "found in key/value store:", val)
		return val
	}

	val, exists = v.defaults[key]
	if exists {
		jww.TRACE.Println(key, "found in defaults:", val)
		return val
	}

	return nil
}

// Check to see if the key has been set in any of the data locations
func IsSet(key string) bool { return v.IsSet(key) }
func (v *viper) IsSet(key string) bool {
	t := v.Get(key)
	return t != nil
}

// Have viper check ENV variables for all
// keys set in config, default & flags
func AutomaticEnv() { v.AutomaticEnv() }
func (v *viper) AutomaticEnv() {
	v.automaticEnvApplied = true
}

// Aliases provide another accessor for the same key.
// This enables one to change a name without breaking the application
func RegisterAlias(alias string, key string) { v.RegisterAlias(alias, key) }
func (v *viper) RegisterAlias(alias string, key string) {
	v.registerAlias(alias, strings.ToLower(key))
}

func (v *viper) registerAlias(alias string, key string) {
	alias = strings.ToLower(alias)
	if alias != key && alias != v.realKey(key) {
		_, exists := v.aliases[alias]

		if !exists {
			// if we alias something that exists in one of the maps to another
			// name, we'll never be able to get that value using the original
			// name, so move the config value to the new realkey.
			if val, ok := v.config[alias]; ok {
				delete(v.config, alias)
				v.config[key] = val
			}
			if val, ok := v.kvstore[alias]; ok {
				delete(v.kvstore, alias)
				v.kvstore[key] = val
			}
			if val, ok := v.defaults[alias]; ok {
				delete(v.defaults, alias)
				v.defaults[key] = val
			}
			if val, ok := v.override[alias]; ok {
				delete(v.override, alias)
				v.override[key] = val
			}
			v.aliases[alias] = key
		}
	} else {
		jww.WARN.Println("Creating circular reference alias", alias, key, v.realKey(key))
	}
}

func (v *viper) realKey(key string) string {
	newkey, exists := v.aliases[key]
	if exists {
		jww.DEBUG.Println("Alias", key, "to", newkey)
		return v.realKey(newkey)
	} else {
		return key
	}
}

// Check to see if the given key (or an alias) is in the config file
func InConfig(key string) bool { return v.InConfig(key) }
func (v *viper) InConfig(key string) bool {
	// if the requested key is an alias, then return the proper key
	key = v.realKey(key)

	_, exists := v.config[key]
	return exists
}

// Set the default value for this key.
// Default only used when no value is provided by the user via flag, config or ENV.
func SetDefault(key string, value interface{}) { v.SetDefault(key, value) }
func (v *viper) SetDefault(key string, value interface{}) {
	// If alias passed in, then set the proper default
	key = v.realKey(strings.ToLower(key))
	v.defaults[key] = value
}

// The user provided value (via flag)
// Will be used instead of values obtained via
// config file, ENV, default, or key/value store
func Set(key string, value interface{}) { v.Set(key, value) }
func (v *viper) Set(key string, value interface{}) {
	// If alias passed in, then set the proper override
	key = v.realKey(strings.ToLower(key))
	v.override[key] = value
}

// Viper will discover and load the configuration file from disk
// and key/value stores, searching in one of the defined paths.
func ReadInConfig() error { return v.ReadInConfig() }
func (v *viper) ReadInConfig() error {
	jww.INFO.Println("Attempting to read in config file")
	if !stringInSlice(v.getConfigType(), SupportedExts) {
		return UnsupportedConfigError(v.getConfigType())
	}

	file, err := ioutil.ReadFile(v.getConfigFile())
	if err != nil {
		return err
	}

	v.marshalReader(bytes.NewReader(file), v.config)
	return nil
}

func ReadRemoteConfig() error { return v.ReadRemoteConfig() }
func (v *viper) ReadRemoteConfig() error {
	err := v.getKeyValueConfig()
	if err != nil {
		return err
	}
	return nil
}

// Marshall a Reader into a map
// Should probably be an unexported function
func marshalReader(in io.Reader, c map[string]interface{}) { v.marshalReader(in, c) }
func (v *viper) marshalReader(in io.Reader, c map[string]interface{}) {
	marshallConfigReader(in, c, v.getConfigType())
}

func (v *viper) insensativiseMaps() {
	insensativiseMap(v.config)
	insensativiseMap(v.defaults)
	insensativiseMap(v.override)
	insensativiseMap(v.kvstore)
}

// retrieve the first found remote configuration
func (v *viper) getKeyValueConfig() error {
	for _, rp := range v.remoteProviders {
		val, err := v.getRemoteConfig(rp)
		if err != nil {
			continue
		}
		v.kvstore = val
		return nil
	}
	return RemoteConfigError("No Files Found")
}

func (v *viper) getRemoteConfig(provider *remoteProvider) (map[string]interface{}, error) {
	var cm crypt.ConfigManager
	var err error

	if provider.secretKeyring != "" {
		kr, err := os.Open(provider.secretKeyring)
		defer kr.Close()
		if err != nil {
			return nil, err
		}
		if provider.provider == "etcd" {
			cm, err = crypt.NewEtcdConfigManager([]string{provider.endpoint}, kr)
		} else {
			cm, err = crypt.NewConsulConfigManager([]string{provider.endpoint}, kr)
		}
	} else {
		if provider.provider == "etcd" {
			cm, err = crypt.NewStandardEtcdConfigManager([]string{provider.endpoint})
		} else {
			cm, err = crypt.NewStandardConsulConfigManager([]string{provider.endpoint})
		}
	}
	if err != nil {
		return nil, err
	}
	b, err := cm.Get(provider.path)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(b)
	v.marshalReader(reader, v.kvstore)
	return v.kvstore, err
}

// Return all keys regardless where they are set
func AllKeys() []string { return v.AllKeys() }
func (v *viper) AllKeys() []string {
	m := map[string]struct{}{}

	for key, _ := range v.defaults {
		m[key] = struct{}{}
	}

	for key, _ := range v.config {
		m[key] = struct{}{}
	}

	for key, _ := range v.kvstore {
		m[key] = struct{}{}
	}

	for key, _ := range v.override {
		m[key] = struct{}{}
	}

	a := []string{}
	for x, _ := range m {
		a = append(a, x)
	}

	return a
}

// Return all settings as a map[string]interface{}
func AllSettings() map[string]interface{} { return v.AllSettings() }
func (v *viper) AllSettings() map[string]interface{} {
	m := map[string]interface{}{}
	for _, x := range v.AllKeys() {
		m[x] = v.Get(x)
	}

	return m
}

// Name for the config file.
// Does not include extension.
func SetConfigName(in string) { v.SetConfigName(in) }
func (v *viper) SetConfigName(in string) {
	if in != "" {
		v.configName = in
	}
}

func SetConfigType(in string) { v.SetConfigType(in) }
func (v *viper) SetConfigType(in string) {
	if in != "" {
		v.configType = in
	}
}

func (v *viper) getConfigType() string {
	if v.configType != "" {
		return v.configType
	}

	cf := v.getConfigFile()
	ext := filepath.Ext(cf)

	if len(ext) > 1 {
		return ext[1:]
	} else {
		return ""
	}
}

func (v *viper) getConfigFile() string {
	// if explicitly set, then use it
	if v.configFile != "" {
		return v.configFile
	}

	cf, err := v.findConfigFile()
	if err != nil {
		return ""
	}

	v.configFile = cf
	return v.getConfigFile()
}

func (v *viper) searchInPath(in string) (filename string) {
	jww.DEBUG.Println("Searching for config in ", in)
	for _, ext := range SupportedExts {
		jww.DEBUG.Println("Checking for", filepath.Join(in, v.configName+"."+ext))
		if b, _ := exists(filepath.Join(in, v.configName+"."+ext)); b {
			jww.DEBUG.Println("Found: ", filepath.Join(in, v.configName+"."+ext))
			return filepath.Join(in, v.configName+"."+ext)
		}
	}

	return ""
}

// search all configPaths for any config file.
// Returns the first path that exists (and is a config file)
func (v *viper) findConfigFile() (string, error) {
	jww.INFO.Println("Searching for config in ", v.configPaths)

	for _, cp := range v.configPaths {
		file := v.searchInPath(cp)
		if file != "" {
			return file, nil
		}
	}

	cwd, _ := findCWD()
	file := v.searchInPath(cwd)
	if file != "" {
		return file, nil
	}

	// try the current working directory
	wd, _ := os.Getwd()
	file = v.searchInPath(wd)
	if file != "" {
		return file, nil
	}
	return "", fmt.Errorf("config file not found in: %s", v.configPaths)
}

func Debug() { v.Debug() }
func (v *viper) Debug() {
	fmt.Println("Config:")
	pretty.Println(v.config)
	fmt.Println("Key/Value Store:")
	pretty.Println(v.kvstore)
	fmt.Println("Env:")
	pretty.Println(v.env)
	fmt.Println("Defaults:")
	pretty.Println(v.defaults)
	fmt.Println("Override:")
	pretty.Println(v.override)
	fmt.Println("Aliases:")
	pretty.Println(v.aliases)
}
