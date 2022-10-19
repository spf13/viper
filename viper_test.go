// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package viper

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil" //nolint:staticcheck
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spf13/viper/internal/testutil"
)

// var yamlExample = []byte(`Hacker: true
// name: steve
// hobbies:
//     - skateboarding
//     - snowboarding
//     - go
// clothing:
//     jacket: leather
//     trousers: denim
//     pants:
//         size: large
// age: 35
// eyes : brown
// beard: true
// `)

var yamlExampleWithExtras = []byte(`Existing: true
Bogus: true
`)

type testUnmarshalExtra struct {
	Existing bool
}

var tomlExample = []byte(`
title = "TOML Example"

[empty.map]

[owner]
organization = "MongoDB"
Bio = "MongoDB Chief Developer Advocate & Hacker at Large"
dob = 1979-05-27T07:32:00Z # First class dates? Why not?`)

var dotenvExample = []byte(`
TITLE_DOTENV="DotEnv Example"
TYPE_DOTENV=donut
NAME_DOTENV=Cake`)

var jsonExample = []byte(`{
"id": "0001",
"type": "donut",
"name": "Cake",
"ppu": 0.55,
"batters": {
        "batter": [
                { "type": "Regular" },
                { "type": "Chocolate" },
                { "type": "Blueberry" },
                { "type": "Devil's Food" }
            ]
    }
}`)

var hclExample = []byte(`
id = "0001"
type = "donut"
name = "Cake"
ppu = 0.55
foos {
	foo {
		key = 1
	}
	foo {
		key = 2
	}
	foo {
		key = 3
	}
	foo {
		key = 4
	}
}`)

var propertiesExample = []byte(`
p_id: 0001
p_type: donut
p_name: Cake
p_ppu: 0.55
p_batters.batter.type: Regular
`)

var remoteExample = []byte(`{
"id":"0002",
"type":"cronut",
"newkey":"remote"
}`)

var iniExample = []byte(`; Package name
NAME        = ini
; Package version
VERSION     = v1
; Package import path
IMPORT_PATH = gopkg.in/%(NAME)s.%(VERSION)s

# Information about package author
# Bio can be written in multiple lines.
[author]
NAME   = Unknown  ; Succeeding comment
E-MAIL = fake@localhost
GITHUB = https://github.com/%(NAME)s
BIO    = """Gopher.
Coding addict.
Good man.
"""  # Succeeding comment`)

func initConfigs() {
	Reset()
	var r io.Reader
	SetConfigType("yaml")
	r = bytes.NewReader(yamlExample)
	unmarshalReader(r, v.config)

	SetConfigType("json")
	r = bytes.NewReader(jsonExample)
	unmarshalReader(r, v.config)

	SetConfigType("hcl")
	r = bytes.NewReader(hclExample)
	unmarshalReader(r, v.config)

	SetConfigType("properties")
	r = bytes.NewReader(propertiesExample)
	unmarshalReader(r, v.config)

	SetConfigType("toml")
	r = bytes.NewReader(tomlExample)
	unmarshalReader(r, v.config)

	SetConfigType("env")
	r = bytes.NewReader(dotenvExample)
	unmarshalReader(r, v.config)

	SetConfigType("json")
	remote := bytes.NewReader(remoteExample)
	unmarshalReader(remote, v.kvstore)

	SetConfigType("ini")
	r = bytes.NewReader(iniExample)
	unmarshalReader(r, v.config)
}

func initConfig(typ, config string) {
	Reset()
	SetConfigType(typ)
	r := strings.NewReader(config)

	if err := unmarshalReader(r, v.config); err != nil {
		panic(err)
	}
}

func initYAML() {
	initConfig("yaml", string(yamlExample))
}

func initJSON() {
	Reset()
	SetConfigType("json")
	r := bytes.NewReader(jsonExample)

	unmarshalReader(r, v.config)
}

func initProperties() {
	Reset()
	SetConfigType("properties")
	r := bytes.NewReader(propertiesExample)

	unmarshalReader(r, v.config)
}

func initTOML() {
	Reset()
	SetConfigType("toml")
	r := bytes.NewReader(tomlExample)

	unmarshalReader(r, v.config)
}

func initDotEnv() {
	Reset()
	SetConfigType("env")
	r := bytes.NewReader(dotenvExample)

	unmarshalReader(r, v.config)
}

func initHcl() {
	Reset()
	SetConfigType("hcl")
	r := bytes.NewReader(hclExample)

	unmarshalReader(r, v.config)
}

func initIni() {
	Reset()
	SetConfigType("ini")
	r := bytes.NewReader(iniExample)

	unmarshalReader(r, v.config)
}

// make directories for testing
func initDirs(t *testing.T) (string, string, func()) {
	var (
		testDirs = []string{`a a`, `b`, `C_`}
		config   = `improbable`
	)

	if runtime.GOOS != "windows" {
		testDirs = append(testDirs, `d\d`)
	}

	root, err := ioutil.TempDir("", "")
	require.NoError(t, err, "Failed to create temporary directory")

	cleanup := true
	defer func() {
		if cleanup {
			os.Chdir("..")
			os.RemoveAll(root)
		}
	}()

	assert.Nil(t, err)

	err = os.Chdir(root)
	require.Nil(t, err)

	for _, dir := range testDirs {
		err = os.Mkdir(dir, 0o750)
		assert.Nil(t, err)

		err = ioutil.WriteFile(
			path.Join(dir, config+".toml"),
			[]byte("key = \"value is "+dir+"\"\n"),
			0o640)
		assert.Nil(t, err)
	}

	cleanup = false
	return root, config, func() {
		os.Chdir("..")
		os.RemoveAll(root)
	}
}

// stubs for PFlag Values
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Type() string {
	return "string"
}

func (s *stringValue) String() string {
	return string(*s)
}

func TestGetConfigFile(t *testing.T) {
	t.Run("config file set", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		_, err = fs.Create(testutil.AbsFilePath(t, "/etc/viper/config.yaml"))
		require.NoError(t, err)

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc/viper")
		v.SetConfigFile(testutil.AbsFilePath(t, "/etc/viper/config.yaml"))

		filename, err := v.getConfigFile()
		assert.Equal(t, testutil.AbsFilePath(t, "/etc/viper/config.yaml"), filename)
		assert.NoError(t, err)
	})

	t.Run("find file", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		_, err = fs.Create(testutil.AbsFilePath(t, "/etc/viper/config.yaml"))
		require.NoError(t, err)

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc/viper")

		filename, err := v.getConfigFile()
		assert.Equal(t, testutil.AbsFilePath(t, "/etc/viper/config.yaml"), filename)
		assert.NoError(t, err)
	})

	t.Run("find files only", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/config"), 0o777)
		require.NoError(t, err)

		_, err = fs.Create(testutil.AbsFilePath(t, "/etc/config/config.yaml"))
		require.NoError(t, err)

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc")
		v.AddConfigPath("/etc/config")

		filename, err := v.getConfigFile()
		assert.Equal(t, testutil.AbsFilePath(t, "/etc/config/config.yaml"), filename)
		assert.NoError(t, err)
	})

	t.Run("precedence", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/home/viper"), 0o777)
		require.NoError(t, err)

		_, err = fs.Create(testutil.AbsFilePath(t, "/home/viper/config.zml"))
		require.NoError(t, err)

		err = fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		_, err = fs.Create(testutil.AbsFilePath(t, "/etc/viper/config.bml"))
		require.NoError(t, err)

		err = fs.Mkdir(testutil.AbsFilePath(t, "/var/viper"), 0o777)
		require.NoError(t, err)

		_, err = fs.Create(testutil.AbsFilePath(t, "/var/viper/config.yaml"))
		require.NoError(t, err)

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/home/viper")
		v.AddConfigPath("/etc/viper")
		v.AddConfigPath("/var/viper")

		filename, err := v.getConfigFile()
		assert.Equal(t, testutil.AbsFilePath(t, "/var/viper/config.yaml"), filename)
		assert.NoError(t, err)
	})

	t.Run("without extension", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		_, err = fs.Create(testutil.AbsFilePath(t, "/etc/viper/.dotfilenoext"))
		require.NoError(t, err)

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc/viper")
		v.SetConfigName(".dotfilenoext")
		v.SetConfigType("yaml")

		filename, err := v.getConfigFile()
		assert.Equal(t, testutil.AbsFilePath(t, "/etc/viper/.dotfilenoext"), filename)
		assert.NoError(t, err)
	})

	t.Run("without extension and config type", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		_, err = fs.Create(testutil.AbsFilePath(t, "/etc/viper/.dotfilenoext"))
		require.NoError(t, err)

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc/viper")
		v.SetConfigName(".dotfilenoext")

		_, err = v.getConfigFile()
		// unless config type is set, files without extension
		// are not considered
		assert.Error(t, err)
	})
}

func TestReadInConfig(t *testing.T) {
	t.Run("config file set", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir("/etc/viper", 0o777)
		require.NoError(t, err)

		file, err := fs.Create(testutil.AbsFilePath(t, "/etc/viper/config.yaml"))
		require.NoError(t, err)

		_, err = file.Write([]byte(`key: value`))
		require.NoError(t, err)

		file.Close()

		v := New()

		v.SetFs(fs)
		v.SetConfigFile(testutil.AbsFilePath(t, "/etc/viper/config.yaml"))

		err = v.ReadInConfig()
		require.NoError(t, err)

		assert.Equal(t, "value", v.Get("key"))
	})

	t.Run("find file", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		file, err := fs.Create(testutil.AbsFilePath(t, "/etc/viper/config.yaml"))
		require.NoError(t, err)

		_, err = file.Write([]byte(`key: value`))
		require.NoError(t, err)

		file.Close()

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc/viper")

		err = v.ReadInConfig()
		require.NoError(t, err)

		assert.Equal(t, "value", v.Get("key"))
	})
}

func TestDefault(t *testing.T) {
	SetDefault("age", 45)
	assert.Equal(t, 45, Get("age"))

	SetDefault("clothing.jacket", "slacks")
	assert.Equal(t, "slacks", Get("clothing.jacket"))

	SetConfigType("yaml")
	err := ReadConfig(bytes.NewBuffer(yamlExample))

	assert.NoError(t, err)
	assert.Equal(t, "leather", Get("clothing.jacket"))
}

func TestUnmarshaling(t *testing.T) {
	SetConfigType("yaml")
	r := bytes.NewReader(yamlExample)

	unmarshalReader(r, v.config)
	assert.True(t, InConfig("name"))
	assert.True(t, InConfig("clothing.jacket"))
	assert.False(t, InConfig("state"))
	assert.False(t, InConfig("clothing.hat"))
	assert.Equal(t, "steve", Get("name"))
	assert.Equal(t, []interface{}{"skateboarding", "snowboarding", "go"}, Get("hobbies"))
	assert.Equal(t, map[string]interface{}{"jacket": "leather", "trousers": "denim", "pants": map[string]interface{}{"size": "large"}}, Get("clothing"))
	assert.Equal(t, 35, Get("age"))
}

func TestUnmarshalExact(t *testing.T) {
	vip := New()
	target := &testUnmarshalExtra{}
	vip.SetConfigType("yaml")
	r := bytes.NewReader(yamlExampleWithExtras)
	vip.ReadConfig(r)
	err := vip.UnmarshalExact(target)
	if err == nil {
		t.Fatal("UnmarshalExact should error when populating a struct from a conf that contains unused fields")
	}
}

func TestOverrides(t *testing.T) {
	Set("age", 40)
	assert.Equal(t, 40, Get("age"))
}

func TestDefaultPost(t *testing.T) {
	assert.NotEqual(t, "NYC", Get("state"))
	SetDefault("state", "NYC")
	assert.Equal(t, "NYC", Get("state"))
}

func TestAliases(t *testing.T) {
	RegisterAlias("years", "age")
	assert.Equal(t, 40, Get("years"))
	Set("years", 45)
	assert.Equal(t, 45, Get("age"))
}

func TestAliasInConfigFile(t *testing.T) {
	// the config file specifies "beard".  If we make this an alias for
	// "hasbeard", we still want the old config file to work with beard.
	RegisterAlias("beard", "hasbeard")
	assert.Equal(t, true, Get("hasbeard"))
	Set("hasbeard", false)
	assert.Equal(t, false, Get("beard"))
}

func TestYML(t *testing.T) {
	initYAML()
	assert.Equal(t, "steve", Get("name"))
}

func TestJSON(t *testing.T) {
	initJSON()
	assert.Equal(t, "0001", Get("id"))
}

func TestProperties(t *testing.T) {
	initProperties()
	assert.Equal(t, "0001", Get("p_id"))
}

func TestTOML(t *testing.T) {
	initTOML()
	assert.Equal(t, "TOML Example", Get("title"))
}

func TestDotEnv(t *testing.T) {
	initDotEnv()
	assert.Equal(t, "DotEnv Example", Get("title_dotenv"))
}

func TestHCL(t *testing.T) {
	initHcl()
	assert.Equal(t, "0001", Get("id"))
	assert.Equal(t, 0.55, Get("ppu"))
	assert.Equal(t, "donut", Get("type"))
	assert.Equal(t, "Cake", Get("name"))
	Set("id", "0002")
	assert.Equal(t, "0002", Get("id"))
	assert.NotEqual(t, "cronut", Get("type"))
}

func TestIni(t *testing.T) {
	initIni()
	assert.Equal(t, "ini", Get("default.name"))
}

func TestRemotePrecedence(t *testing.T) {
	initJSON()

	remote := bytes.NewReader(remoteExample)
	assert.Equal(t, "0001", Get("id"))
	unmarshalReader(remote, v.kvstore)
	assert.Equal(t, "0001", Get("id"))
	assert.NotEqual(t, "cronut", Get("type"))
	assert.Equal(t, "remote", Get("newkey"))
	Set("newkey", "newvalue")
	assert.NotEqual(t, "remote", Get("newkey"))
	assert.Equal(t, "newvalue", Get("newkey"))
	Set("newkey", "remote")
}

func TestEnv(t *testing.T) {
	initJSON()

	BindEnv("id")
	BindEnv("f", "FOOD", "OLD_FOOD")

	testutil.Setenv(t, "ID", "13")
	testutil.Setenv(t, "FOOD", "apple")
	testutil.Setenv(t, "OLD_FOOD", "banana")
	testutil.Setenv(t, "NAME", "crunk")

	assert.Equal(t, "13", Get("id"))
	assert.Equal(t, "apple", Get("f"))
	assert.Equal(t, "Cake", Get("name"))

	AutomaticEnv()

	assert.Equal(t, "crunk", Get("name"))
}

func TestMultipleEnv(t *testing.T) {
	initJSON()

	BindEnv("f", "FOOD", "OLD_FOOD")

	testutil.Setenv(t, "OLD_FOOD", "banana")

	assert.Equal(t, "banana", Get("f"))
}

func TestEmptyEnv(t *testing.T) {
	initJSON()

	BindEnv("type") // Empty environment variable
	BindEnv("name") // Bound, but not set environment variable

	testutil.Setenv(t, "TYPE", "")

	assert.Equal(t, "donut", Get("type"))
	assert.Equal(t, "Cake", Get("name"))
}

func TestEmptyEnv_Allowed(t *testing.T) {
	initJSON()

	AllowEmptyEnv(true)

	BindEnv("type") // Empty environment variable
	BindEnv("name") // Bound, but not set environment variable

	testutil.Setenv(t, "TYPE", "")

	assert.Equal(t, "", Get("type"))
	assert.Equal(t, "Cake", Get("name"))
}

func TestEmptyMap_Allowed(t *testing.T) {
	initTOML()
	AllowEmptyMap(true)

	allkeys := sort.StringSlice(AllKeys())
	allkeys.Sort()

	assert.Equal(t, sort.StringSlice(sort.StringSlice{"empty.map", "owner.bio", "owner.dob", "owner.organization", "title"}), allkeys)
}

func TestEnvPrefix(t *testing.T) {
	initJSON()

	SetEnvPrefix("foo") // will be uppercased automatically
	BindEnv("id")
	BindEnv("f", "FOOD") // not using prefix

	testutil.Setenv(t, "FOO_ID", "13")
	testutil.Setenv(t, "FOOD", "apple")
	testutil.Setenv(t, "FOO_NAME", "crunk")

	assert.Equal(t, "13", Get("id"))
	assert.Equal(t, "apple", Get("f"))
	assert.Equal(t, "Cake", Get("name"))

	AutomaticEnv()

	assert.Equal(t, "crunk", Get("name"))
}

func TestAutoEnv(t *testing.T) {
	Reset()

	AutomaticEnv()

	testutil.Setenv(t, "FOO_BAR", "13")

	assert.Equal(t, "13", Get("foo_bar"))
}

func TestAutoEnvWithPrefix(t *testing.T) {
	Reset()

	AutomaticEnv()
	SetEnvPrefix("Baz")

	testutil.Setenv(t, "BAZ_BAR", "13")

	assert.Equal(t, "13", Get("bar"))
}

func TestSetEnvKeyReplacer(t *testing.T) {
	Reset()

	AutomaticEnv()

	testutil.Setenv(t, "REFRESH_INTERVAL", "30s")

	replacer := strings.NewReplacer("-", "_")
	SetEnvKeyReplacer(replacer)

	assert.Equal(t, "30s", Get("refresh-interval"))
}

func TestEnvKeyReplacer(t *testing.T) {
	v := NewWithOptions(EnvKeyReplacer(strings.NewReplacer("-", "_")))

	v.AutomaticEnv()

	testutil.Setenv(t, "REFRESH_INTERVAL", "30s")

	assert.Equal(t, "30s", v.Get("refresh-interval"))
}

func TestAllKeys(t *testing.T) {
	initConfigs()

	ks := sort.StringSlice{
		"title",
		"author.bio",
		"author.e-mail",
		"author.github",
		"author.name",
		"newkey",
		"owner.organization",
		"owner.dob",
		"owner.bio",
		"name",
		"beard",
		"ppu",
		"batters.batter",
		"hobbies",
		"clothing.jacket",
		"clothing.trousers",
		"default.import_path",
		"default.name",
		"default.version",
		"clothing.pants.size",
		"age",
		"hacker",
		"id",
		"type",
		"eyes",
		"p_id",
		"p_ppu",
		"p_batters.batter.type",
		"p_type",
		"p_name",
		"foos",
		"title_dotenv",
		"type_dotenv",
		"name_dotenv",
	}
	dob, _ := time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	all := map[string]interface{}{
		"owner": map[string]interface{}{
			"organization": "MongoDB",
			"bio":          "MongoDB Chief Developer Advocate & Hacker at Large",
			"dob":          dob,
		},
		"title": "TOML Example",
		"author": map[string]interface{}{
			"e-mail": "fake@localhost",
			"github": "https://github.com/Unknown",
			"name":   "Unknown",
			"bio":    "Gopher.\nCoding addict.\nGood man.\n",
		},
		"ppu":  0.55,
		"eyes": "brown",
		"clothing": map[string]interface{}{
			"trousers": "denim",
			"jacket":   "leather",
			"pants":    map[string]interface{}{"size": "large"},
		},
		"default": map[string]interface{}{
			"import_path": "gopkg.in/ini.v1",
			"name":        "ini",
			"version":     "v1",
		},
		"id": "0001",
		"batters": map[string]interface{}{
			"batter": []interface{}{
				map[string]interface{}{"type": "Regular"},
				map[string]interface{}{"type": "Chocolate"},
				map[string]interface{}{"type": "Blueberry"},
				map[string]interface{}{"type": "Devil's Food"},
			},
		},
		"hacker": true,
		"beard":  true,
		"hobbies": []interface{}{
			"skateboarding",
			"snowboarding",
			"go",
		},
		"age":    35,
		"type":   "donut",
		"newkey": "remote",
		"name":   "Cake",
		"p_id":   "0001",
		"p_ppu":  "0.55",
		"p_name": "Cake",
		"p_batters": map[string]interface{}{
			"batter": map[string]interface{}{"type": "Regular"},
		},
		"p_type": "donut",
		"foos": []map[string]interface{}{
			{
				"foo": []map[string]interface{}{
					{"key": 1},
					{"key": 2},
					{"key": 3},
					{"key": 4},
				},
			},
		},
		"title_dotenv": "DotEnv Example",
		"type_dotenv":  "donut",
		"name_dotenv":  "Cake",
	}

	allkeys := sort.StringSlice(AllKeys())
	allkeys.Sort()
	ks.Sort()

	assert.Equal(t, ks, allkeys)
	assert.Equal(t, all, AllSettings())
}

func TestAllKeysWithEnv(t *testing.T) {
	v := New()

	// bind and define environment variables (including a nested one)
	v.BindEnv("id")
	v.BindEnv("foo.bar")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	testutil.Setenv(t, "ID", "13")
	testutil.Setenv(t, "FOO_BAR", "baz")

	expectedKeys := sort.StringSlice{"id", "foo.bar"}
	expectedKeys.Sort()
	keys := sort.StringSlice(v.AllKeys())
	keys.Sort()
	assert.Equal(t, expectedKeys, keys)
}

func TestAliasesOfAliases(t *testing.T) {
	Set("Title", "Checking Case")
	RegisterAlias("Foo", "Bar")
	RegisterAlias("Bar", "Title")
	assert.Equal(t, "Checking Case", Get("FOO"))
}

func TestRecursiveAliases(t *testing.T) {
	RegisterAlias("Baz", "Roo")
	RegisterAlias("Roo", "baz")
}

func TestUnmarshal(t *testing.T) {
	SetDefault("port", 1313)
	Set("name", "Steve")
	Set("duration", "1s1ms")
	Set("modes", []int{1, 2, 3})

	type config struct {
		Port     int
		Name     string
		Duration time.Duration
		Modes    []int
	}

	var C config

	err := Unmarshal(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	assert.Equal(
		t,
		&config{
			Name:     "Steve",
			Port:     1313,
			Duration: time.Second + time.Millisecond,
			Modes:    []int{1, 2, 3},
		},
		&C,
	)

	Set("port", 1234)
	err = Unmarshal(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	assert.Equal(
		t,
		&config{
			Name:     "Steve",
			Port:     1234,
			Duration: time.Second + time.Millisecond,
			Modes:    []int{1, 2, 3},
		},
		&C,
	)
}

func TestUnmarshalWithDecoderOptions(t *testing.T) {
	Set("credentials", "{\"foo\":\"bar\"}")

	opt := DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		// Custom Decode Hook Function
		func(rf reflect.Kind, rt reflect.Kind, data interface{}) (interface{}, error) {
			if rf != reflect.String || rt != reflect.Map {
				return data, nil
			}
			m := map[string]string{}
			raw := data.(string)
			if raw == "" {
				return m, nil
			}
			return m, json.Unmarshal([]byte(raw), &m)
		},
	))

	type config struct {
		Credentials map[string]string
	}

	var C config

	err := Unmarshal(&C, opt)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	assert.Equal(t, &config{
		Credentials: map[string]string{"foo": "bar"},
	}, &C)
}

func TestBindPFlags(t *testing.T) {
	v := New() // create independent Viper object
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)

	testValues := map[string]*string{
		"host":     nil,
		"port":     nil,
		"endpoint": nil,
	}

	mutatedTestValues := map[string]string{
		"host":     "localhost",
		"port":     "6060",
		"endpoint": "/public",
	}

	for name := range testValues {
		testValues[name] = flagSet.String(name, "", "test")
	}

	err := v.BindPFlags(flagSet)
	if err != nil {
		t.Fatalf("error binding flag set, %v", err)
	}

	flagSet.VisitAll(func(flag *pflag.Flag) {
		flag.Value.Set(mutatedTestValues[flag.Name])
		flag.Changed = true
	})

	for name, expected := range mutatedTestValues {
		assert.Equal(t, expected, v.Get(name))
	}
}

//nolint:dupl
func TestBindPFlagsStringSlice(t *testing.T) {
	tests := []struct {
		Expected []string
		Value    string
	}{
		{[]string{}, ""},
		{[]string{"jeden"}, "jeden"},
		{[]string{"dwa", "trzy"}, "dwa,trzy"},
		{[]string{"cztery", "piec , szesc"}, "cztery,\"piec , szesc\""},
	}

	v := New() // create independent Viper object
	defaultVal := []string{"default"}
	v.SetDefault("stringslice", defaultVal)

	for _, testValue := range tests {
		flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
		flagSet.StringSlice("stringslice", testValue.Expected, "test")

		for _, changed := range []bool{true, false} {
			flagSet.VisitAll(func(f *pflag.Flag) {
				f.Value.Set(testValue.Value)
				f.Changed = changed
			})

			err := v.BindPFlags(flagSet)
			if err != nil {
				t.Fatalf("error binding flag set, %v", err)
			}

			type TestStr struct {
				StringSlice []string
			}
			val := &TestStr{}
			if err := v.Unmarshal(val); err != nil {
				t.Fatalf("%+#v cannot unmarshal: %s", testValue.Value, err)
			}
			if changed {
				assert.Equal(t, testValue.Expected, val.StringSlice)
				assert.Equal(t, testValue.Expected, v.Get("stringslice"))
			} else {
				assert.Equal(t, defaultVal, val.StringSlice)
			}
		}
	}
}

//nolint:dupl
func TestBindPFlagsStringArray(t *testing.T) {
	tests := []struct {
		Expected []string
		Value    string
	}{
		{[]string{}, ""},
		{[]string{"jeden"}, "jeden"},
		{[]string{"dwa,trzy"}, "dwa,trzy"},
		{[]string{"cztery,\"piec , szesc\""}, "cztery,\"piec , szesc\""},
	}

	v := New() // create independent Viper object
	defaultVal := []string{"default"}
	v.SetDefault("stringarray", defaultVal)

	for _, testValue := range tests {
		flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
		flagSet.StringArray("stringarray", testValue.Expected, "test")

		for _, changed := range []bool{true, false} {
			flagSet.VisitAll(func(f *pflag.Flag) {
				f.Value.Set(testValue.Value)
				f.Changed = changed
			})

			err := v.BindPFlags(flagSet)
			if err != nil {
				t.Fatalf("error binding flag set, %v", err)
			}

			type TestStr struct {
				StringArray []string
			}
			val := &TestStr{}
			if err := v.Unmarshal(val); err != nil {
				t.Fatalf("%+#v cannot unmarshal: %s", testValue.Value, err)
			}
			if changed {
				assert.Equal(t, testValue.Expected, val.StringArray)
				assert.Equal(t, testValue.Expected, v.Get("stringarray"))
			} else {
				assert.Equal(t, defaultVal, val.StringArray)
			}
		}
	}
}

//nolint:dupl
func TestBindPFlagsIntSlice(t *testing.T) {
	tests := []struct {
		Expected []int
		Value    string
	}{
		{[]int{}, ""},
		{[]int{1}, "1"},
		{[]int{2, 3}, "2,3"},
	}

	v := New() // create independent Viper object
	defaultVal := []int{0}
	v.SetDefault("intslice", defaultVal)

	for _, testValue := range tests {
		flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
		flagSet.IntSlice("intslice", testValue.Expected, "test")

		for _, changed := range []bool{true, false} {
			flagSet.VisitAll(func(f *pflag.Flag) {
				f.Value.Set(testValue.Value)
				f.Changed = changed
			})

			err := v.BindPFlags(flagSet)
			if err != nil {
				t.Fatalf("error binding flag set, %v", err)
			}

			type TestInt struct {
				IntSlice []int
			}
			val := &TestInt{}
			if err := v.Unmarshal(val); err != nil {
				t.Fatalf("%+#v cannot unmarshal: %s", testValue.Value, err)
			}
			if changed {
				assert.Equal(t, testValue.Expected, val.IntSlice)
				assert.Equal(t, testValue.Expected, v.Get("intslice"))
			} else {
				assert.Equal(t, defaultVal, val.IntSlice)
			}
		}
	}
}

func TestBindPFlag(t *testing.T) {
	testString := "testing"
	testValue := newStringValue(testString, &testString)

	flag := &pflag.Flag{
		Name:    "testflag",
		Value:   testValue,
		Changed: false,
	}

	BindPFlag("testvalue", flag)

	assert.Equal(t, testString, Get("testvalue"))

	flag.Value.Set("testing_mutate")
	flag.Changed = true // hack for pflag usage

	assert.Equal(t, "testing_mutate", Get("testvalue"))
}

func TestBindPFlagDetectNilFlag(t *testing.T) {
	result := BindPFlag("testvalue", nil)
	assert.Error(t, result)
}

func TestBindPFlagStringToString(t *testing.T) {
	tests := []struct {
		Expected map[string]string
		Value    string
	}{
		{map[string]string{}, ""},
		{map[string]string{"yo": "hi"}, "yo=hi"},
		{map[string]string{"yo": "hi", "oh": "hi=there"}, "yo=hi,oh=hi=there"},
		{map[string]string{"yo": ""}, "yo="},
		{map[string]string{"yo": "", "oh": "hi=there"}, "yo=,oh=hi=there"},
	}

	v := New() // create independent Viper object
	defaultVal := map[string]string{}
	v.SetDefault("stringtostring", defaultVal)

	for _, testValue := range tests {
		flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
		flagSet.StringToString("stringtostring", testValue.Expected, "test")

		for _, changed := range []bool{true, false} {
			flagSet.VisitAll(func(f *pflag.Flag) {
				f.Value.Set(testValue.Value)
				f.Changed = changed
			})

			err := v.BindPFlags(flagSet)
			if err != nil {
				t.Fatalf("error binding flag set, %v", err)
			}

			type TestMap struct {
				StringToString map[string]string
			}
			val := &TestMap{}
			if err := v.Unmarshal(val); err != nil {
				t.Fatalf("%+#v cannot unmarshal: %s", testValue.Value, err)
			}
			if changed {
				assert.Equal(t, testValue.Expected, val.StringToString)
			} else {
				assert.Equal(t, defaultVal, val.StringToString)
			}
		}
	}
}

func TestBoundCaseSensitivity(t *testing.T) {
	assert.Equal(t, "brown", Get("eyes"))

	BindEnv("eYEs", "TURTLE_EYES")

	testutil.Setenv(t, "TURTLE_EYES", "blue")

	assert.Equal(t, "blue", Get("eyes"))

	testString := "green"
	testValue := newStringValue(testString, &testString)

	flag := &pflag.Flag{
		Name:    "eyeballs",
		Value:   testValue,
		Changed: true,
	}

	BindPFlag("eYEs", flag)
	assert.Equal(t, "green", Get("eyes"))
}

func TestSizeInBytes(t *testing.T) {
	input := map[string]uint{
		"":               0,
		"b":              0,
		"12 bytes":       0,
		"200000000000gb": 0,
		"12 b":           12,
		"43 MB":          43 * (1 << 20),
		"10mb":           10 * (1 << 20),
		"1gb":            1 << 30,
	}

	for str, expected := range input {
		assert.Equal(t, expected, parseSizeInBytes(str), str)
	}
}

func TestFindsNestedKeys(t *testing.T) {
	initConfigs()
	dob, _ := time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")

	Set("super", map[string]interface{}{
		"deep": map[string]interface{}{
			"nested": "value",
		},
	})

	expected := map[string]interface{}{
		"super": map[string]interface{}{
			"deep": map[string]interface{}{
				"nested": "value",
			},
		},
		"super.deep": map[string]interface{}{
			"nested": "value",
		},
		"super.deep.nested":  "value",
		"owner.organization": "MongoDB",
		"batters.batter": []interface{}{
			map[string]interface{}{
				"type": "Regular",
			},
			map[string]interface{}{
				"type": "Chocolate",
			},
			map[string]interface{}{
				"type": "Blueberry",
			},
			map[string]interface{}{
				"type": "Devil's Food",
			},
		},
		"hobbies": []interface{}{
			"skateboarding", "snowboarding", "go",
		},
		"TITLE_DOTENV": "DotEnv Example",
		"TYPE_DOTENV":  "donut",
		"NAME_DOTENV":  "Cake",
		"title":        "TOML Example",
		"newkey":       "remote",
		"batters": map[string]interface{}{
			"batter": []interface{}{
				map[string]interface{}{
					"type": "Regular",
				},
				map[string]interface{}{
					"type": "Chocolate",
				},
				map[string]interface{}{
					"type": "Blueberry",
				},
				map[string]interface{}{
					"type": "Devil's Food",
				},
			},
		},
		"eyes": "brown",
		"age":  35,
		"owner": map[string]interface{}{
			"organization": "MongoDB",
			"bio":          "MongoDB Chief Developer Advocate & Hacker at Large",
			"dob":          dob,
		},
		"owner.bio": "MongoDB Chief Developer Advocate & Hacker at Large",
		"type":      "donut",
		"id":        "0001",
		"name":      "Cake",
		"hacker":    true,
		"ppu":       0.55,
		"clothing": map[string]interface{}{
			"jacket":   "leather",
			"trousers": "denim",
			"pants": map[string]interface{}{
				"size": "large",
			},
		},
		"clothing.jacket":     "leather",
		"clothing.pants.size": "large",
		"clothing.trousers":   "denim",
		"owner.dob":           dob,
		"beard":               true,
		"foos": []map[string]interface{}{
			{
				"foo": []map[string]interface{}{
					{
						"key": 1,
					},
					{
						"key": 2,
					},
					{
						"key": 3,
					},
					{
						"key": 4,
					},
				},
			},
		},
	}

	for key, expectedValue := range expected {
		assert.Equal(t, expectedValue, v.Get(key))
	}
}

func TestReadBufConfig(t *testing.T) {
	v := New()
	v.SetConfigType("yaml")
	v.ReadConfig(bytes.NewBuffer(yamlExample))
	t.Log(v.AllKeys())

	assert.True(t, v.InConfig("name"))
	assert.True(t, v.InConfig("clothing.jacket"))
	assert.False(t, v.InConfig("state"))
	assert.False(t, v.InConfig("clothing.hat"))
	assert.Equal(t, "steve", v.Get("name"))
	assert.Equal(t, []interface{}{"skateboarding", "snowboarding", "go"}, v.Get("hobbies"))
	assert.Equal(t, map[string]interface{}{"jacket": "leather", "trousers": "denim", "pants": map[string]interface{}{"size": "large"}}, v.Get("clothing"))
	assert.Equal(t, 35, v.Get("age"))
}

func TestIsSet(t *testing.T) {
	v := New()
	v.SetConfigType("yaml")

	/* config and defaults */
	v.ReadConfig(bytes.NewBuffer(yamlExample))
	v.SetDefault("clothing.shoes", "sneakers")

	assert.True(t, v.IsSet("clothing"))
	assert.True(t, v.IsSet("clothing.jacket"))
	assert.False(t, v.IsSet("clothing.jackets"))
	assert.True(t, v.IsSet("clothing.shoes"))

	/* state change */
	assert.False(t, v.IsSet("helloworld"))
	v.Set("helloworld", "fubar")
	assert.True(t, v.IsSet("helloworld"))

	/* env */
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.BindEnv("eyes")
	v.BindEnv("foo")
	v.BindEnv("clothing.hat")
	v.BindEnv("clothing.hats")

	testutil.Setenv(t, "FOO", "bar")
	testutil.Setenv(t, "CLOTHING_HAT", "bowler")

	assert.True(t, v.IsSet("eyes"))           // in the config file
	assert.True(t, v.IsSet("foo"))            // in the environment
	assert.True(t, v.IsSet("clothing.hat"))   // in the environment
	assert.False(t, v.IsSet("clothing.hats")) // not defined

	/* flags */
	flagset := pflag.NewFlagSet("testisset", pflag.ContinueOnError)
	flagset.Bool("foobaz", false, "foobaz")
	flagset.Bool("barbaz", false, "barbaz")
	foobaz, barbaz := flagset.Lookup("foobaz"), flagset.Lookup("barbaz")
	v.BindPFlag("foobaz", foobaz)
	v.BindPFlag("barbaz", barbaz)
	barbaz.Value.Set("true")
	barbaz.Changed = true // hack for pflag usage

	assert.False(t, v.IsSet("foobaz"))
	assert.True(t, v.IsSet("barbaz"))
}

func TestDirsSearch(t *testing.T) {
	root, config, cleanup := initDirs(t)
	defer cleanup()

	v := New()
	v.SetConfigName(config)
	v.SetDefault(`key`, `default`)

	entries, err := ioutil.ReadDir(root)
	assert.Nil(t, err)
	for _, e := range entries {
		if e.IsDir() {
			v.AddConfigPath(e.Name())
		}
	}

	err = v.ReadInConfig()
	assert.Nil(t, err)

	assert.Equal(t, `value is `+filepath.Base(v.configPaths[0]), v.GetString(`key`))
}

func TestWrongDirsSearchNotFound(t *testing.T) {
	_, config, cleanup := initDirs(t)
	defer cleanup()

	v := New()
	v.SetConfigName(config)
	v.SetDefault(`key`, `default`)

	v.AddConfigPath(`whattayoutalkingbout`)
	v.AddConfigPath(`thispathaintthere`)

	err := v.ReadInConfig()
	assert.Equal(t, reflect.TypeOf(ConfigFileNotFoundError{"", ""}), reflect.TypeOf(err))

	// Even though config did not load and the error might have
	// been ignored by the client, the default still loads
	assert.Equal(t, `default`, v.GetString(`key`))
}

func TestWrongDirsSearchNotFoundForMerge(t *testing.T) {
	_, config, cleanup := initDirs(t)
	defer cleanup()

	v := New()
	v.SetConfigName(config)
	v.SetDefault(`key`, `default`)

	v.AddConfigPath(`whattayoutalkingbout`)
	v.AddConfigPath(`thispathaintthere`)

	err := v.MergeInConfig()
	assert.Equal(t, reflect.TypeOf(ConfigFileNotFoundError{"", ""}), reflect.TypeOf(err))

	// Even though config did not load and the error might have
	// been ignored by the client, the default still loads
	assert.Equal(t, `default`, v.GetString(`key`))
}

func TestSub(t *testing.T) {
	v := New()
	v.SetConfigType("yaml")
	v.ReadConfig(bytes.NewBuffer(yamlExample))

	subv := v.Sub("clothing")
	assert.Equal(t, v.Get("clothing.pants.size"), subv.Get("pants.size"))

	subv = v.Sub("clothing.pants")
	assert.Equal(t, v.Get("clothing.pants.size"), subv.Get("size"))

	subv = v.Sub("clothing.pants.size")
	assert.Equal(t, (*Viper)(nil), subv)

	subv = v.Sub("missing.key")
	assert.Equal(t, (*Viper)(nil), subv)
}

var hclWriteExpected = []byte(`"foos" = {
  "foo" = {
    "key" = 1
  }

  "foo" = {
    "key" = 2
  }

  "foo" = {
    "key" = 3
  }

  "foo" = {
    "key" = 4
  }
}

"id" = "0001"

"name" = "Cake"

"ppu" = 0.55

"type" = "donut"`)

var jsonWriteExpected = []byte(`{
  "batters": {
    "batter": [
      {
        "type": "Regular"
      },
      {
        "type": "Chocolate"
      },
      {
        "type": "Blueberry"
      },
      {
        "type": "Devil's Food"
      }
    ]
  },
  "id": "0001",
  "name": "Cake",
  "ppu": 0.55,
  "type": "donut"
}`)

var propertiesWriteExpected = []byte(`p_id = 0001
p_type = donut
p_name = Cake
p_ppu = 0.55
p_batters.batter.type = Regular
`)

// var yamlWriteExpected = []byte(`age: 35
// beard: true
// clothing:
//     jacket: leather
//     pants:
//         size: large
//     trousers: denim
// eyes: brown
// hacker: true
// hobbies:
//     - skateboarding
//     - snowboarding
//     - go
// name: steve
// `)

func TestWriteConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	testCases := map[string]struct {
		configName      string
		inConfigType    string
		outConfigType   string
		fileName        string
		input           []byte
		expectedContent []byte
	}{
		"hcl with file extension": {
			configName:      "c",
			inConfigType:    "hcl",
			outConfigType:   "hcl",
			fileName:        "c.hcl",
			input:           hclExample,
			expectedContent: hclWriteExpected,
		},
		"hcl without file extension": {
			configName:      "c",
			inConfigType:    "hcl",
			outConfigType:   "hcl",
			fileName:        "c",
			input:           hclExample,
			expectedContent: hclWriteExpected,
		},
		"hcl with file extension and mismatch type": {
			configName:      "c",
			inConfigType:    "hcl",
			outConfigType:   "json",
			fileName:        "c.hcl",
			input:           hclExample,
			expectedContent: hclWriteExpected,
		},
		"json with file extension": {
			configName:      "c",
			inConfigType:    "json",
			outConfigType:   "json",
			fileName:        "c.json",
			input:           jsonExample,
			expectedContent: jsonWriteExpected,
		},
		"json without file extension": {
			configName:      "c",
			inConfigType:    "json",
			outConfigType:   "json",
			fileName:        "c",
			input:           jsonExample,
			expectedContent: jsonWriteExpected,
		},
		"json with file extension and mismatch type": {
			configName:      "c",
			inConfigType:    "json",
			outConfigType:   "hcl",
			fileName:        "c.json",
			input:           jsonExample,
			expectedContent: jsonWriteExpected,
		},
		"properties with file extension": {
			configName:      "c",
			inConfigType:    "properties",
			outConfigType:   "properties",
			fileName:        "c.properties",
			input:           propertiesExample,
			expectedContent: propertiesWriteExpected,
		},
		"properties without file extension": {
			configName:      "c",
			inConfigType:    "properties",
			outConfigType:   "properties",
			fileName:        "c",
			input:           propertiesExample,
			expectedContent: propertiesWriteExpected,
		},
		"yaml with file extension": {
			configName:      "c",
			inConfigType:    "yaml",
			outConfigType:   "yaml",
			fileName:        "c.yaml",
			input:           yamlExample,
			expectedContent: yamlWriteExpected,
		},
		"yaml without file extension": {
			configName:      "c",
			inConfigType:    "yaml",
			outConfigType:   "yaml",
			fileName:        "c",
			input:           yamlExample,
			expectedContent: yamlWriteExpected,
		},
		"yaml with file extension and mismatch type": {
			configName:      "c",
			inConfigType:    "yaml",
			outConfigType:   "json",
			fileName:        "c.yaml",
			input:           yamlExample,
			expectedContent: yamlWriteExpected,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			v := New()
			v.SetFs(fs)
			v.SetConfigName(tc.fileName)
			v.SetConfigType(tc.inConfigType)

			err := v.ReadConfig(bytes.NewBuffer(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			v.SetConfigType(tc.outConfigType)
			if err := v.WriteConfigAs(tc.fileName); err != nil {
				t.Fatal(err)
			}
			read, err := afero.ReadFile(fs, tc.fileName)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expectedContent, read)
		})
	}
}

func TestWriteConfigTOML(t *testing.T) {
	fs := afero.NewMemMapFs()

	testCases := map[string]struct {
		configName string
		configType string
		fileName   string
		input      []byte
	}{
		"with file extension": {
			configName: "c",
			configType: "toml",
			fileName:   "c.toml",
			input:      tomlExample,
		},
		"without file extension": {
			configName: "c",
			configType: "toml",
			fileName:   "c",
			input:      tomlExample,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			v := New()
			v.SetFs(fs)
			v.SetConfigName(tc.configName)
			v.SetConfigType(tc.configType)
			err := v.ReadConfig(bytes.NewBuffer(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			if err := v.WriteConfigAs(tc.fileName); err != nil {
				t.Fatal(err)
			}

			// The TOML String method does not order the contents.
			// Therefore, we must read the generated file and compare the data.
			v2 := New()
			v2.SetFs(fs)
			v2.SetConfigName(tc.configName)
			v2.SetConfigType(tc.configType)
			v2.SetConfigFile(tc.fileName)
			err = v2.ReadInConfig()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, v.GetString("title"), v2.GetString("title"))
			assert.Equal(t, v.GetString("owner.bio"), v2.GetString("owner.bio"))
			assert.Equal(t, v.GetString("owner.dob"), v2.GetString("owner.dob"))
			assert.Equal(t, v.GetString("owner.organization"), v2.GetString("owner.organization"))
		})
	}
}

func TestWriteConfigDotEnv(t *testing.T) {
	fs := afero.NewMemMapFs()
	testCases := map[string]struct {
		configName string
		configType string
		fileName   string
		input      []byte
	}{
		"with file extension": {
			configName: "c",
			configType: "env",
			fileName:   "c.env",
			input:      dotenvExample,
		},
		"without file extension": {
			configName: "c",
			configType: "env",
			fileName:   "c",
			input:      dotenvExample,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			v := New()
			v.SetFs(fs)
			v.SetConfigName(tc.configName)
			v.SetConfigType(tc.configType)
			err := v.ReadConfig(bytes.NewBuffer(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			if err := v.WriteConfigAs(tc.fileName); err != nil {
				t.Fatal(err)
			}

			// The TOML String method does not order the contents.
			// Therefore, we must read the generated file and compare the data.
			v2 := New()
			v2.SetFs(fs)
			v2.SetConfigName(tc.configName)
			v2.SetConfigType(tc.configType)
			v2.SetConfigFile(tc.fileName)
			err = v2.ReadInConfig()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, v.GetString("title_dotenv"), v2.GetString("title_dotenv"))
			assert.Equal(t, v.GetString("type_dotenv"), v2.GetString("type_dotenv"))
			assert.Equal(t, v.GetString("kind_dotenv"), v2.GetString("kind_dotenv"))
		})
	}
}

func TestSafeWriteConfig(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	v.SetFs(fs)
	v.AddConfigPath("/test")
	v.SetConfigName("c")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(bytes.NewBuffer(yamlExample)))
	require.NoError(t, v.SafeWriteConfig())
	read, err := afero.ReadFile(fs, testutil.AbsFilePath(t, "/test/c.yaml"))
	require.NoError(t, err)
	assert.Equal(t, yamlWriteExpected, read)
}

func TestSafeWriteConfigWithMissingConfigPath(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	v.SetFs(fs)
	v.SetConfigName("c")
	v.SetConfigType("yaml")
	require.EqualError(t, v.SafeWriteConfig(), "missing configuration for 'configPath'")
}

func TestSafeWriteConfigWithExistingFile(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	fs.Create(testutil.AbsFilePath(t, "/test/c.yaml"))
	v.SetFs(fs)
	v.AddConfigPath("/test")
	v.SetConfigName("c")
	v.SetConfigType("yaml")
	err := v.SafeWriteConfig()
	require.Error(t, err)
	_, ok := err.(ConfigFileAlreadyExistsError)
	assert.True(t, ok, "Expected ConfigFileAlreadyExistsError")
}

func TestSafeWriteAsConfig(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	v.SetFs(fs)
	err := v.ReadConfig(bytes.NewBuffer(yamlExample))
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, v.SafeWriteConfigAs("/test/c.yaml"))
	if _, err = afero.ReadFile(fs, "/test/c.yaml"); err != nil {
		t.Fatal(err)
	}
}

func TestSafeWriteConfigAsWithExistingFile(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	fs.Create("/test/c.yaml")
	v.SetFs(fs)
	err := v.SafeWriteConfigAs("/test/c.yaml")
	require.Error(t, err)
	_, ok := err.(ConfigFileAlreadyExistsError)
	assert.True(t, ok, "Expected ConfigFileAlreadyExistsError")
}

func TestWriteHiddenFile(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	fs.Create(testutil.AbsFilePath(t, "/test/.config"))
	v.SetFs(fs)

	v.SetConfigName(".config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/test")

	err := v.ReadInConfig()
	require.NoError(t, err)

	err = v.WriteConfig()
	require.NoError(t, err)
}

var yamlMergeExampleTgt = []byte(`
hello:
    pop: 37890
    largenum: 765432101234567
    num2pow63: 9223372036854775808
    universe: null
    world:
    - us
    - uk
    - fr
    - de
`)

var yamlMergeExampleSrc = []byte(`
hello:
    pop: 45000
    largenum: 7654321001234567
    universe:
    - mw
    - ad
    ints:
    - 1
    - 2
fu: bar
`)

var jsonMergeExampleTgt = []byte(`
{
	"hello": {
		"foo": null,
		"pop": 123456
	}
}
`)

var jsonMergeExampleSrc = []byte(`
{
	"hello": {
		"foo": "foo str",
		"pop": "pop str"
	}
}
`)

func TestMergeConfig(t *testing.T) {
	v := New()
	v.SetConfigType("yml")
	if err := v.ReadConfig(bytes.NewBuffer(yamlMergeExampleTgt)); err != nil {
		t.Fatal(err)
	}

	if pop := v.GetInt("hello.pop"); pop != 37890 {
		t.Fatalf("pop != 37890, = %d", pop)
	}

	if pop := v.GetInt32("hello.pop"); pop != int32(37890) {
		t.Fatalf("pop != 37890, = %d", pop)
	}

	if pop := v.GetInt64("hello.largenum"); pop != int64(765432101234567) {
		t.Fatalf("int64 largenum != 765432101234567, = %d", pop)
	}

	if pop := v.GetUint("hello.pop"); pop != 37890 {
		t.Fatalf("uint pop != 37890, = %d", pop)
	}

	if pop := v.GetUint16("hello.pop"); pop != uint16(37890) {
		t.Fatalf("uint pop != 37890, = %d", pop)
	}

	if pop := v.GetUint32("hello.pop"); pop != 37890 {
		t.Fatalf("uint32 pop != 37890, = %d", pop)
	}

	if pop := v.GetUint64("hello.num2pow63"); pop != 9223372036854775808 {
		t.Fatalf("uint64 num2pow63 != 9223372036854775808, = %d", pop)
	}

	if world := v.GetStringSlice("hello.world"); len(world) != 4 {
		t.Fatalf("len(world) != 4, = %d", len(world))
	}

	if fu := v.GetString("fu"); fu != "" {
		t.Fatalf("fu != \"\", = %s", fu)
	}

	if err := v.MergeConfig(bytes.NewBuffer(yamlMergeExampleSrc)); err != nil {
		t.Fatal(err)
	}

	if pop := v.GetInt("hello.pop"); pop != 45000 {
		t.Fatalf("pop != 45000, = %d", pop)
	}

	if pop := v.GetInt32("hello.pop"); pop != int32(45000) {
		t.Fatalf("pop != 45000, = %d", pop)
	}

	if pop := v.GetInt64("hello.largenum"); pop != int64(7654321001234567) {
		t.Fatalf("int64 largenum != 7654321001234567, = %d", pop)
	}

	if world := v.GetStringSlice("hello.world"); len(world) != 4 {
		t.Fatalf("len(world) != 4, = %d", len(world))
	}

	if universe := v.GetStringSlice("hello.universe"); len(universe) != 2 {
		t.Fatalf("len(universe) != 2, = %d", len(universe))
	}

	if ints := v.GetIntSlice("hello.ints"); len(ints) != 2 {
		t.Fatalf("len(ints) != 2, = %d", len(ints))
	}

	if fu := v.GetString("fu"); fu != "bar" {
		t.Fatalf("fu != \"bar\", = %s", fu)
	}
}

func TestMergeConfigOverrideType(t *testing.T) {
	v := New()
	v.SetConfigType("json")
	if err := v.ReadConfig(bytes.NewBuffer(jsonMergeExampleTgt)); err != nil {
		t.Fatal(err)
	}

	if err := v.MergeConfig(bytes.NewBuffer(jsonMergeExampleSrc)); err != nil {
		t.Fatal(err)
	}

	if pop := v.GetString("hello.pop"); pop != "pop str" {
		t.Fatalf("pop != \"pop str\", = %s", pop)
	}

	if foo := v.GetString("hello.foo"); foo != "foo str" {
		t.Fatalf("foo != \"foo str\", = %s", foo)
	}
}

func TestMergeConfigNoMerge(t *testing.T) {
	v := New()
	v.SetConfigType("yml")
	if err := v.ReadConfig(bytes.NewBuffer(yamlMergeExampleTgt)); err != nil {
		t.Fatal(err)
	}

	if pop := v.GetInt("hello.pop"); pop != 37890 {
		t.Fatalf("pop != 37890, = %d", pop)
	}

	if world := v.GetStringSlice("hello.world"); len(world) != 4 {
		t.Fatalf("len(world) != 4, = %d", len(world))
	}

	if fu := v.GetString("fu"); fu != "" {
		t.Fatalf("fu != \"\", = %s", fu)
	}

	if err := v.ReadConfig(bytes.NewBuffer(yamlMergeExampleSrc)); err != nil {
		t.Fatal(err)
	}

	if pop := v.GetInt("hello.pop"); pop != 45000 {
		t.Fatalf("pop != 45000, = %d", pop)
	}

	if world := v.GetStringSlice("hello.world"); len(world) != 0 {
		t.Fatalf("len(world) != 0, = %d", len(world))
	}

	if universe := v.GetStringSlice("hello.universe"); len(universe) != 2 {
		t.Fatalf("len(universe) != 2, = %d", len(universe))
	}

	if ints := v.GetIntSlice("hello.ints"); len(ints) != 2 {
		t.Fatalf("len(ints) != 2, = %d", len(ints))
	}

	if fu := v.GetString("fu"); fu != "bar" {
		t.Fatalf("fu != \"bar\", = %s", fu)
	}
}

func TestMergeConfigMap(t *testing.T) {
	v := New()
	v.SetConfigType("yml")
	if err := v.ReadConfig(bytes.NewBuffer(yamlMergeExampleTgt)); err != nil {
		t.Fatal(err)
	}

	assert := func(i int) {
		large := v.GetInt64("hello.largenum")
		pop := v.GetInt("hello.pop")
		if large != 765432101234567 {
			t.Fatal("Got large num:", large)
		}

		if pop != i {
			t.Fatal("Got pop:", pop)
		}
	}

	assert(37890)

	update := map[string]interface{}{
		"Hello": map[string]interface{}{
			"Pop": 1234,
		},
		"World": map[interface{}]interface{}{
			"Rock": 345,
		},
	}

	if err := v.MergeConfigMap(update); err != nil {
		t.Fatal(err)
	}

	if rock := v.GetInt("world.rock"); rock != 345 {
		t.Fatal("Got rock:", rock)
	}

	assert(1234)
}

func TestUnmarshalingWithAliases(t *testing.T) {
	v := New()
	v.SetDefault("ID", 1)
	v.Set("name", "Steve")
	v.Set("lastname", "Owen")

	v.RegisterAlias("UserID", "ID")
	v.RegisterAlias("Firstname", "name")
	v.RegisterAlias("Surname", "lastname")

	type config struct {
		ID        int
		FirstName string
		Surname   string
	}

	var C config
	err := v.Unmarshal(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	assert.Equal(t, &config{ID: 1, FirstName: "Steve", Surname: "Owen"}, &C)
}

func TestSetConfigNameClearsFileCache(t *testing.T) {
	SetConfigFile("/tmp/config.yaml")
	SetConfigName("default")
	f, err := v.getConfigFile()
	if err == nil {
		t.Fatalf("config file cache should have been cleared")
	}
	assert.Empty(t, f)
}

func TestShadowedNestedValue(t *testing.T) {
	config := `name: steve
clothing:
  jacket: leather
  trousers: denim
  pants:
    size: large
`
	initConfig("yaml", config)

	assert.Equal(t, "steve", GetString("name"))

	polyester := "polyester"
	SetDefault("clothing.shirt", polyester)
	SetDefault("clothing.jacket.price", 100)

	assert.Equal(t, "leather", GetString("clothing.jacket"))
	assert.Nil(t, Get("clothing.jacket.price"))
	assert.Equal(t, polyester, GetString("clothing.shirt"))

	clothingSettings := AllSettings()["clothing"].(map[string]interface{})
	assert.Equal(t, "leather", clothingSettings["jacket"])
	assert.Equal(t, polyester, clothingSettings["shirt"])
}

func TestDotParameter(t *testing.T) {
	initJSON()
	// shoud take precedence over batters defined in jsonExample
	r := bytes.NewReader([]byte(`{ "batters.batter": [ { "type": "Small" } ] }`))
	unmarshalReader(r, v.config)

	actual := Get("batters.batter")
	expected := []interface{}{map[string]interface{}{"type": "Small"}}
	assert.Equal(t, expected, actual)
}

func TestCaseInsensitive(t *testing.T) {
	for _, config := range []struct {
		typ     string
		content string
	}{
		{"yaml", `
aBcD: 1
eF:
  gH: 2
  iJk: 3
  Lm:
    nO: 4
    P:
      Q: 5
      R: 6
`},
		{"json", `{
  "aBcD": 1,
  "eF": {
    "iJk": 3,
    "Lm": {
      "P": {
        "Q": 5,
        "R": 6
      },
      "nO": 4
    },
    "gH": 2
  }
}`},
		{"toml", `aBcD = 1
[eF]
gH = 2
iJk = 3
[eF.Lm]
nO = 4
[eF.Lm.P]
Q = 5
R = 6
`},
	} {
		doTestCaseInsensitive(t, config.typ, config.content)
	}
}

func TestCaseInsensitiveSet(t *testing.T) {
	Reset()
	m1 := map[string]interface{}{
		"Foo": 32,
		"Bar": map[interface{}]interface{}{
			"ABc": "A",
			"cDE": "B",
		},
	}

	m2 := map[string]interface{}{
		"Foo": 52,
		"Bar": map[interface{}]interface{}{
			"bCd": "A",
			"eFG": "B",
		},
	}

	Set("Given1", m1)
	Set("Number1", 42)

	SetDefault("Given2", m2)
	SetDefault("Number2", 52)

	// Verify SetDefault
	if v := Get("number2"); v != 52 {
		t.Fatalf("Expected 52 got %q", v)
	}

	if v := Get("given2.foo"); v != 52 {
		t.Fatalf("Expected 52 got %q", v)
	}

	if v := Get("given2.bar.bcd"); v != "A" {
		t.Fatalf("Expected A got %q", v)
	}

	if _, ok := m2["Foo"]; !ok {
		t.Fatal("Input map changed")
	}

	// Verify Set
	if v := Get("number1"); v != 42 {
		t.Fatalf("Expected 42 got %q", v)
	}

	if v := Get("given1.foo"); v != 32 {
		t.Fatalf("Expected 32 got %q", v)
	}

	if v := Get("given1.bar.abc"); v != "A" {
		t.Fatalf("Expected A got %q", v)
	}

	if _, ok := m1["Foo"]; !ok {
		t.Fatal("Input map changed")
	}
}

func TestParseNested(t *testing.T) {
	type duration struct {
		Delay time.Duration
	}

	type item struct {
		Name   string
		Delay  time.Duration
		Nested duration
	}

	config := `[[parent]]
	delay="100ms"
	[parent.nested]
	delay="200ms"
`
	initConfig("toml", config)

	var items []item
	err := v.UnmarshalKey("parent", &items)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	assert.Equal(t, 1, len(items))
	assert.Equal(t, 100*time.Millisecond, items[0].Delay)
	assert.Equal(t, 200*time.Millisecond, items[0].Nested.Delay)
}

func doTestCaseInsensitive(t *testing.T, typ, config string) {
	initConfig(typ, config)
	Set("RfD", true)
	assert.Equal(t, true, Get("rfd"))
	assert.Equal(t, true, Get("rFD"))
	assert.Equal(t, 1, cast.ToInt(Get("abcd")))
	assert.Equal(t, 1, cast.ToInt(Get("Abcd")))
	assert.Equal(t, 2, cast.ToInt(Get("ef.gh")))
	assert.Equal(t, 3, cast.ToInt(Get("ef.ijk")))
	assert.Equal(t, 4, cast.ToInt(Get("ef.lm.no")))
	assert.Equal(t, 5, cast.ToInt(Get("ef.lm.p.q")))
}

func newViperWithConfigFile(t *testing.T) (*Viper, string, func()) {
	watchDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	configFile := path.Join(watchDir, "config.yaml")
	err = ioutil.WriteFile(configFile, []byte("foo: bar\n"), 0o640)
	require.Nil(t, err)
	cleanup := func() {
		os.RemoveAll(watchDir)
	}
	v := New()
	v.SetConfigFile(configFile)
	err = v.ReadInConfig()
	require.Nil(t, err)
	require.Equal(t, "bar", v.Get("foo"))
	return v, configFile, cleanup
}

func newViperWithSymlinkedConfigFile(t *testing.T) (*Viper, string, string, func()) {
	watchDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	dataDir1 := path.Join(watchDir, "data1")
	err = os.Mkdir(dataDir1, 0o777)
	require.Nil(t, err)
	realConfigFile := path.Join(dataDir1, "config.yaml")
	t.Logf("Real config file location: %s\n", realConfigFile)
	err = ioutil.WriteFile(realConfigFile, []byte("foo: bar\n"), 0o640)
	require.Nil(t, err)
	cleanup := func() {
		os.RemoveAll(watchDir)
	}
	// now, symlink the tm `data1` dir to `data` in the baseDir
	os.Symlink(dataDir1, path.Join(watchDir, "data"))
	// and link the `<watchdir>/datadir1/config.yaml` to `<watchdir>/config.yaml`
	configFile := path.Join(watchDir, "config.yaml")
	os.Symlink(path.Join(watchDir, "data", "config.yaml"), configFile)
	t.Logf("Config file location: %s\n", path.Join(watchDir, "config.yaml"))
	// init Viper
	v := New()
	v.SetConfigFile(configFile)
	err = v.ReadInConfig()
	require.Nil(t, err)
	require.Equal(t, "bar", v.Get("foo"))
	return v, watchDir, configFile, cleanup
}

func TestWatchFile(t *testing.T) {
	if runtime.GOOS == "linux" {
		// TODO(bep) FIX ME
		t.Skip("Skip test on Linux ...")
	}

	t.Run("file content changed", func(t *testing.T) {
		// given a `config.yaml` file being watched
		v, configFile, cleanup := newViperWithConfigFile(t)
		defer cleanup()
		_, err := os.Stat(configFile)
		require.NoError(t, err)
		t.Logf("test config file: %s\n", configFile)
		wg := sync.WaitGroup{}
		wg.Add(1)
		var wgDoneOnce sync.Once // OnConfigChange is called twice on Windows
		v.OnConfigChange(func(in fsnotify.Event) {
			t.Logf("config file changed")
			wgDoneOnce.Do(func() {
				wg.Done()
			})
		})
		v.WatchConfig()
		// when overwriting the file and waiting for the custom change notification handler to be triggered
		err = ioutil.WriteFile(configFile, []byte("foo: baz\n"), 0o640)
		wg.Wait()
		// then the config value should have changed
		require.Nil(t, err)
		assert.Equal(t, "baz", v.Get("foo"))
	})

	t.Run("link to real file changed (à la Kubernetes)", func(t *testing.T) {
		// skip if not executed on Linux
		if runtime.GOOS != "linux" {
			t.Skipf("Skipping test as symlink replacements don't work on non-linux environment...")
		}
		v, watchDir, _, _ := newViperWithSymlinkedConfigFile(t)
		// defer cleanup()
		wg := sync.WaitGroup{}
		v.WatchConfig()
		v.OnConfigChange(func(in fsnotify.Event) {
			t.Logf("config file changed")
			wg.Done()
		})
		wg.Add(1)
		// when link to another `config.yaml` file
		dataDir2 := path.Join(watchDir, "data2")
		err := os.Mkdir(dataDir2, 0o777)
		require.Nil(t, err)
		configFile2 := path.Join(dataDir2, "config.yaml")
		err = ioutil.WriteFile(configFile2, []byte("foo: baz\n"), 0o640)
		require.Nil(t, err)
		// change the symlink using the `ln -sfn` command
		err = exec.Command("ln", "-sfn", dataDir2, path.Join(watchDir, "data")).Run()
		require.Nil(t, err)
		wg.Wait()
		// then
		require.Nil(t, err)
		assert.Equal(t, "baz", v.Get("foo"))
	})
}

func TestUnmarshal_DotSeparatorBackwardCompatibility(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("foo.bar", "cobra_flag", "")

	v := New()
	assert.NoError(t, v.BindPFlags(flags))

	config := &struct {
		Foo struct {
			Bar string
		}
	}{}

	assert.NoError(t, v.Unmarshal(config))
	assert.Equal(t, "cobra_flag", config.Foo.Bar)
}

// var yamlExampleWithDot = []byte(`Hacker: true
// name: steve
// hobbies:
//     - skateboarding
//     - snowboarding
//     - go
// clothing:
//     jacket: leather
//     trousers: denim
//     pants:
//         size: large
// age: 35
// eyes : brown
// beard: true
// emails:
//     steve@hacker.com:
//         created: 01/02/03
//         active: true
// `)

func TestKeyDelimiter(t *testing.T) {
	v := NewWithOptions(KeyDelimiter("::"))
	v.SetConfigType("yaml")
	r := strings.NewReader(string(yamlExampleWithDot))

	err := v.unmarshalReader(r, v.config)
	require.NoError(t, err)

	values := map[string]interface{}{
		"image": map[string]interface{}{
			"repository": "someImage",
			"tag":        "1.0.0",
		},
		"ingress": map[string]interface{}{
			"annotations": map[string]interface{}{
				"traefik.frontend.rule.type":                 "PathPrefix",
				"traefik.ingress.kubernetes.io/ssl-redirect": "true",
			},
		},
	}

	v.SetDefault("charts::values", values)

	assert.Equal(t, "leather", v.GetString("clothing::jacket"))
	assert.Equal(t, "01/02/03", v.GetString("emails::steve@hacker.com::created"))

	type config struct {
		Charts struct {
			Values map[string]interface{}
		}
	}

	expected := config{
		Charts: struct {
			Values map[string]interface{}
		}{
			Values: values,
		},
	}

	var actual config

	assert.NoError(t, v.Unmarshal(&actual))

	assert.Equal(t, expected, actual)
}

var yamlDeepNestedSlices = []byte(`TV:
- title: "The Expanse"
  title_i18n:
    USA: "The Expanse"
    Japan: "エクスパンス -巨獣めざめる-"
  seasons:
  - first_released: "December 14, 2015"
    episodes:
    - title: "Dulcinea"
      air_date: "December 14, 2015"
    - title: "The Big Empty"
      air_date: "December 15, 2015"
    - title: "Remember the Cant"
      air_date: "December 22, 2015"
  - first_released: "February 1, 2017"
    episodes:
    - title: "Safe"
      air_date: "February 1, 2017"
    - title: "Doors & Corners"
      air_date: "February 1, 2017"
    - title: "Static"
      air_date: "February 8, 2017"
  episodes:
    - ["Dulcinea", "The Big Empty", "Remember the Cant"]
    - ["Safe", "Doors & Corners", "Static"]
`)

func TestSliceIndexAccess(t *testing.T) {
	v.SetConfigType("yaml")
	r := strings.NewReader(string(yamlDeepNestedSlices))

	err := v.unmarshalReader(r, v.config)
	require.NoError(t, err)

	assert.Equal(t, "The Expanse", v.GetString("tv.0.title"))
	assert.Equal(t, "February 1, 2017", v.GetString("tv.0.seasons.1.first_released"))
	assert.Equal(t, "Static", v.GetString("tv.0.seasons.1.episodes.2.title"))
	assert.Equal(t, "December 15, 2015", v.GetString("tv.0.seasons.0.episodes.1.air_date"))

	// Test nested keys with capital letters
	assert.Equal(t, "The Expanse", v.GetString("tv.0.title_i18n.USA"))
	assert.Equal(t, "エクスパンス -巨獣めざめる-", v.GetString("tv.0.title_i18n.Japan"))

	// Test for index out of bounds
	assert.Equal(t, "", v.GetString("tv.0.seasons.2.first_released"))

	// Accessing multidimensional arrays
	assert.Equal(t, "Static", v.GetString("tv.0.episodes.1.2"))
}

func BenchmarkGetBool(b *testing.B) {
	key := "BenchmarkGetBool"
	v = New()
	v.Set(key, true)

	for i := 0; i < b.N; i++ {
		if !v.GetBool(key) {
			b.Fatal("GetBool returned false")
		}
	}
}

func BenchmarkGet(b *testing.B) {
	key := "BenchmarkGet"
	v = New()
	v.Set(key, true)

	for i := 0; i < b.N; i++ {
		if !v.Get(key).(bool) {
			b.Fatal("Get returned false")
		}
	}
}

// BenchmarkGetBoolFromMap is the "perfect result" for the above.
func BenchmarkGetBoolFromMap(b *testing.B) {
	m := make(map[string]bool)
	key := "BenchmarkGetBool"
	m[key] = true

	for i := 0; i < b.N; i++ {
		if !m[key] {
			b.Fatal("Map value was false")
		}
	}
}

// Skip some tests on Windows that kept failing when Windows was added to the CI as a target.
func skipWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skip test on Windows")
	}
}
