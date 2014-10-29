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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/kr/pretty"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/pflag"
	crypt "github.com/xordataexchange/crypt/config"
	"gopkg.in/yaml.v1"
)

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

// A set of paths to look for the config file in
var configPaths []string

// A set of remote providers to search for the configuration
var remoteProviders []*remoteProvider

// Name of file to look for inside the path
var configName string = "config"

// extensions Supported
var SupportedExts []string = []string{"json", "toml", "yaml", "yml"}
var SupportedRemoteProviders []string = []string{"etcd", "consul"}
var configFile string
var configType string

var config map[string]interface{} = make(map[string]interface{})
var override map[string]interface{} = make(map[string]interface{})
var env map[string]string = make(map[string]string)
var defaults map[string]interface{} = make(map[string]interface{})
var kvstore map[string]interface{} = make(map[string]interface{})
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

// AddRemoteProvider adds a remote configuration source.
// Remote Providers are searched in the order they are added.
// provider is a string value, "etcd" or "consul" are currently supported.
// endpoint is the url.  etcd requires http://ip:port  consul requires ip:port
// path is the path in the k/v store to retrieve configuration
// To retrieve a config file called myapp.json from /configs/myapp.json
// you should set path to /configs and set config name (SetConfigName()) to
// "myapp"
func AddRemoteProvider(provider, endpoint, path string) error {
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
		if !providerPathExists(rp) {
			remoteProviders = append(remoteProviders, rp)
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
		if !providerPathExists(rp) {
			remoteProviders = append(remoteProviders, rp)
		}
	}
	return nil
}

func providerPathExists(p *remoteProvider) bool {

	for _, y := range remoteProviders {
		if reflect.DeepEqual(y, p) {
			return true
		}
	}
	return false
}

type UnsupportedRemoteProviderError string

func (str UnsupportedRemoteProviderError) Error() string {
	return fmt.Sprintf("Unsupported Remote Provider Type %q", string(str))
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
	err = mapstructure.Decode(kvstore, rawVal)
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
	pflags[strings.ToLower(key)] = flag

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
func BindEnv(input ...string) (err error) {
	var key, envkey string
	if len(input) == 0 {
		return fmt.Errorf("BindEnv missing key to bind to")
	}

	key = strings.ToLower(input[0])

	if len(input) == 1 {
		envkey = strings.ToUpper(key)
	} else {
		envkey = input[1]
	}

	env[key] = envkey

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

	envkey, exists := env[key]
	if exists {
		jww.TRACE.Println(key, "registered as env var", envkey)
		if val = os.Getenv(envkey); val != "" {
			jww.TRACE.Println(envkey, "found in environement with val:", val)
			return val
		} else {
			jww.TRACE.Println(envkey, "env value unset:")
		}
	}

	val, exists = config[key]
	if exists {
		jww.TRACE.Println(key, "found in config:", val)
		return val
	}

	val, exists = kvstore[key]
	if exists {
		jww.TRACE.Println(key, "found in key/value store:", val)
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

// Have viper check ENV variables for all
// keys set in config, default & flags
func AutomaticEnv() {
	for _, x := range AllKeys() {
		BindEnv(x)
	}
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
			if val, ok := kvstore[alias]; ok {
				delete(kvstore, alias)
				kvstore[key] = val
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
// Will be used instead of values obtained via
// config file, ENV, default, or key/value store
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
// and key/value stores, searching in one of the defined paths.
func ReadInConfig() error {
	jww.INFO.Println("Attempting to read in config file")
	if !stringInSlice(getConfigType(), SupportedExts) {
		return UnsupportedConfigError(getConfigType())
	}

	file, err := ioutil.ReadFile(getConfigFile())
	if err != nil {
		return err
	}

	MarshallReader(bytes.NewReader(file), config)
	return nil
}
func ReadRemoteConfig() error {
	err := getKeyValueConfig()
	if err != nil {
		return err
	}
	return nil
}
func MarshallReader(in io.Reader, c map[string]interface{}) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	switch getConfigType() {
	case "yaml", "yml":
		if err := yaml.Unmarshal(buf.Bytes(), &c); err != nil {
			jww.ERROR.Fatalf("Error parsing config: %s", err)
		}

	case "json":
		if err := json.Unmarshal(buf.Bytes(), &c); err != nil {
			jww.ERROR.Fatalf("Error parsing config: %s", err)
		}

	case "toml":
		if _, err := toml.Decode(buf.String(), &c); err != nil {
			jww.ERROR.Fatalf("Error parsing config: %s", err)
		}
	}

	insensativiseMap(c)
}

func insensativiseMaps() {
	insensativiseMap(config)
	insensativiseMap(defaults)
	insensativiseMap(override)
	insensativiseMap(kvstore)
}

// retrieve the first found remote configuration
func getKeyValueConfig() error {
	for _, rp := range remoteProviders {
		val, err := getRemoteConfig(rp)
		if err != nil {
			continue
		}
		kvstore = val
		return nil
	}
	return RemoteConfigError("No Files Found")
}

type RemoteConfigError string

func (rce RemoteConfigError) Error() string {
	return fmt.Sprintf("Remote Configurations Error: %s", string(rce))
}

func getRemoteConfig(provider *remoteProvider) (map[string]interface{}, error) {
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
	MarshallReader(reader, kvstore)
	return kvstore, err
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

func AllKeys() []string {
	m := map[string]struct{}{}

	for key, _ := range defaults {
		m[key] = struct{}{}
	}

	for key, _ := range config {
		m[key] = struct{}{}
	}

	for key, _ := range kvstore {
		m[key] = struct{}{}
	}

	for key, _ := range override {
		m[key] = struct{}{}
	}

	a := []string{}
	for x, _ := range m {
		a = append(a, x)
	}

	return a
}

func AllSettings() map[string]interface{} {
	m := map[string]interface{}{}
	for _, x := range AllKeys() {
		m[x] = Get(x)
	}

	return m
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
	fmt.Println("Key/Value Store:")
	pretty.Println(kvstore)
	fmt.Println("Env:")
	pretty.Println(env)
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

	kvstore = make(map[string]interface{})
	config = make(map[string]interface{})
	override = make(map[string]interface{})
	env = make(map[string]string)
	defaults = make(map[string]interface{})
	aliases = make(map[string]string)
}
