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
	"os/exec"
	"path"
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
)

var yamlExample = []byte(`Hacker: true
name: steve
hobbies:
- skateboarding
- snowboarding
- go
clothing:
  jacket: leather
  trousers: denim
  pants:
    size: large
age: 35
eyes : brown
beard: true
`)

var yamlExampleWithExtras = []byte(`Existing: true
Bogus: true
`)

type testUnmarshalExtra struct {
	Existing bool
}

var tomlExample = []byte(`
title = "TOML Example"

[owner]
organization = "MongoDB"
Bio = "MongoDB Chief Developer Advocate & Hacker at Large"
dob = 1979-05-27T07:32:00Z # First class dates? Why not?`)

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

func initConfigs(v *Viper) {
	var r io.Reader
	v.SetConfigType("yaml")
	r = bytes.NewReader(yamlExample)
	v.unmarshalReader(r, v.config)

	v.SetConfigType("json")
	r = bytes.NewReader(jsonExample)
	v.unmarshalReader(r, v.config)

	v.SetConfigType("hcl")
	r = bytes.NewReader(hclExample)
	v.unmarshalReader(r, v.config)

	v.SetConfigType("properties")
	r = bytes.NewReader(propertiesExample)
	v.unmarshalReader(r, v.config)

	v.SetConfigType("toml")
	r = bytes.NewReader(tomlExample)
	v.unmarshalReader(r, v.config)

	v.SetConfigType("json")
	remote := bytes.NewReader(remoteExample)
	v.unmarshalReader(remote, v.kvstore)
}

func initConfig(v *Viper, typ, config string) {
	v.SetConfigType(typ)
	r := strings.NewReader(config)

	if err := v.unmarshalReader(r, v.config); err != nil {
		panic(err)
	}
}

func initYAML(v *Viper) {
	initConfig(v, "yaml", string(yamlExample))
}

func initJSON(v *Viper) {
	v.SetConfigType("json")
	r := bytes.NewReader(jsonExample)

	v.unmarshalReader(r, v.config)
}

func initProperties(v *Viper) {
	v.SetConfigType("properties")
	r := bytes.NewReader(propertiesExample)

	v.unmarshalReader(r, v.config)
}

func initTOML(v *Viper) {
	v.SetConfigType("toml")
	r := bytes.NewReader(tomlExample)

	v.unmarshalReader(r, v.config)
}

func initHcl(v *Viper) {
	v.SetConfigType("hcl")
	r := bytes.NewReader(hclExample)

	v.unmarshalReader(r, v.config)
}

// make directories for testing
func initDirs(t *testing.T) (string, string, func()) {

	var (
		testDirs = []string{`a a`, `b`, `c\c`, `D_`}
		config   = `improbable`
	)

	root, err := ioutil.TempDir("", "")

	cleanup := true
	defer func() {
		if cleanup {
			os.Chdir("..")
			os.RemoveAll(root)
		}
	}()

	assert.Nil(t, err)

	err = os.Chdir(root)
	assert.Nil(t, err)

	for _, dir := range testDirs {
		err = os.Mkdir(dir, 0750)
		assert.Nil(t, err)

		err = ioutil.WriteFile(
			path.Join(dir, config+".toml"),
			[]byte("key = \"value is "+dir+"\"\n"),
			0640)
		assert.Nil(t, err)
	}

	cleanup = false
	return root, config, func() {
		os.Chdir("..")
		os.RemoveAll(root)
	}
}

//stubs for PFlag Values
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
	return fmt.Sprintf("%s", *s)
}

func TestBasics(t *testing.T) {
	v := New()
	v.SetConfigFile("/tmp/config.yaml")
	filename, err := v.getConfigFile()
	assert.Equal(t, "/tmp/config.yaml", filename)
	assert.NoError(t, err)
}

func TestDefault(t *testing.T) {
	v := New()

	v.SetDefault("age", 45)
	assert.Equal(t, 45, v.Get("age"))

	v.SetDefault("clothing.jacket", "slacks")
	assert.Equal(t, "slacks", v.Get("clothing.jacket"))

	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(yamlExample))

	assert.NoError(t, err)
	assert.Equal(t, "leather", v.Get("clothing.jacket"))
}

func TestUnmarshaling(t *testing.T) {
	v := New()

	v.SetConfigType("yaml")
	r := bytes.NewReader(yamlExample)

	v.unmarshalReader(r, v.config)
	assert.True(t, v.InConfig("name"))
	assert.False(t, v.InConfig("state"))
	assert.Equal(t, "steve", v.Get("name"))
	assert.Equal(t, []interface{}{"skateboarding", "snowboarding", "go"}, v.Get("hobbies"))
	assert.Equal(t, map[string]interface{}{"jacket": "leather", "trousers": "denim", "pants": map[string]interface{}{"size": "large"}}, v.Get("clothing"))
	assert.Equal(t, 35, v.Get("age"))
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
	v := New()
	v.Set("age", 40)
	assert.Equal(t, 40, v.Get("age"))
}

func TestDefaultPost(t *testing.T) {
	v := New()
	assert.NotEqual(t, "NYC", v.Get("state"))
	v.SetDefault("state", "NYC")
	assert.Equal(t, "NYC", v.Get("state"))
}

func TestAliases(t *testing.T) {
	v := New()
	initYAML(v)
	v.RegisterAlias("years", "age")
	assert.Equal(t, 35, v.Get("years"))
	v.Set("years", 45)
	assert.Equal(t, 45, v.Get("age"))
}

func TestAliasInConfigFile(t *testing.T) {
	v := New()
	initYAML(v)
	// the config file specifies "beard".  If we make this an alias for
	// "hasbeard", we still want the old config file to work with beard.
	v.RegisterAlias("beard", "hasbeard")
	assert.Equal(t, true, v.Get("hasbeard"))
	v.Set("hasbeard", false)
	assert.Equal(t, false, v.Get("beard"))
}

func TestYML(t *testing.T) {
	v := New()
	initYAML(v)
	assert.Equal(t, "steve", v.Get("name"))
}

func TestJSON(t *testing.T) {
	v := New()
	initJSON(v)
	assert.Equal(t, "0001", v.Get("id"))
}

func TestProperties(t *testing.T) {
	v := New()
	initProperties(v)
	assert.Equal(t, "0001", v.Get("p_id"))
}

func TestTOML(t *testing.T) {
	v := New()
	initTOML(v)
	assert.Equal(t, "TOML Example", v.Get("title"))
}

func TestHCL(t *testing.T) {
	v := New()
	initHcl(v)
	assert.Equal(t, "0001", v.Get("id"))
	assert.Equal(t, 0.55, v.Get("ppu"))
	assert.Equal(t, "donut", v.Get("type"))
	assert.Equal(t, "Cake", v.Get("name"))
	v.Set("id", "0002")
	assert.Equal(t, "0002", v.Get("id"))
	assert.NotEqual(t, "cronut", v.Get("type"))
}

func TestRemotePrecedence(t *testing.T) {
	v := New()
	initJSON(v)
	remote := bytes.NewReader(remoteExample)
	assert.Equal(t, "0001", v.Get("id"))
	v.unmarshalReader(remote, v.kvstore)
	assert.Equal(t, "0001", v.Get("id"))
	assert.NotEqual(t, "cronut", v.Get("type"))
	assert.Equal(t, "remote", v.Get("newkey"))
	v.Set("newkey", "newvalue")
	assert.NotEqual(t, "remote", v.Get("newkey"))
	assert.Equal(t, "newvalue", v.Get("newkey"))
	v.Set("newkey", "remote")
}

func TestEnv(t *testing.T) {
	v := New()
	initJSON(v)

	v.BindEnv("id")
	v.BindEnv("f", "FOOD", "DEPRECATED_FOOD")

	os.Setenv("ID", "13")
	os.Setenv("FOOD", "apple")
	os.Setenv("DEPRECATED_FOOD", "banana")
	os.Setenv("NAME", "crunk")

	assert.Equal(t, "13", v.Get("id"))
	assert.Equal(t, "apple", v.Get("f"))
	assert.Equal(t, "Cake", v.Get("name"))

	v.AutomaticEnv()

	assert.Equal(t, "crunk", v.Get("name"))
}

func TestEnvTransformer(t *testing.T) {
	v := New()
	initJSON(v)

	v.BindEnv("id", "MY_ID")
	v.SetEnvKeyTransformer("id", func(in string) interface{} { return "transformed" })
	os.Setenv("MY_ID", "14")

	assert.Equal(t, "transformed", v.Get("id"))
}

func TestMultipleEnv(t *testing.T) {
	v := New()
	initJSON(v)

	v.BindEnv("id")
	v.BindEnv("f", "FOOD", "DEPRECATED_FOOD")

	os.Setenv("ID", "13")
	os.Unsetenv("FOOD")
	os.Setenv("DEPRECATED_FOOD", "banana")
	os.Setenv("NAME", "crunk")

	assert.Equal(t, "13", v.Get("id"))
	assert.Equal(t, "banana", v.Get("f"))
	assert.Equal(t, "Cake", v.Get("name"))

	v.AutomaticEnv()

	assert.Equal(t, "crunk", v.Get("name"))
}

func TestEmptyEnv(t *testing.T) {
	v := New()
	initJSON(v)

	v.BindEnv("type") // Empty environment variable
	v.BindEnv("name") // Bound, but not set environment variable

	os.Clearenv()

	os.Setenv("TYPE", "")

	assert.Equal(t, "donut", v.Get("type"))
	assert.Equal(t, "Cake", v.Get("name"))
}

func TestEmptyEnv_Allowed(t *testing.T) {
	v := New()
	initJSON(v)

	v.AllowEmptyEnv(true)

	v.BindEnv("type") // Empty environment variable
	v.BindEnv("name") // Bound, but not set environment variable

	os.Clearenv()

	os.Setenv("TYPE", "")

	assert.Equal(t, "", v.Get("type"))
	assert.Equal(t, "Cake", v.Get("name"))
}

func TestEnvPrefix(t *testing.T) {
	v := New()
	initJSON(v)

	v.SetEnvPrefix("foo") // will be uppercased automatically
	v.BindEnv("id")
	v.BindEnv("f", "FOOD") // not using prefix

	os.Setenv("FOO_ID", "13")
	os.Setenv("FOOD", "apple")
	os.Setenv("FOO_NAME", "crunk")

	assert.Equal(t, "13", v.Get("id"))
	assert.Equal(t, "apple", v.Get("f"))
	assert.Equal(t, "Cake", v.Get("name"))

	v.AutomaticEnv()

	assert.Equal(t, "crunk", v.Get("name"))
}

func TestAutoEnv(t *testing.T) {
	v := New()
	v.AutomaticEnv()
	os.Setenv("FOO_BAR", "13")
	assert.Equal(t, "13", v.Get("foo_bar"))
}

func TestAutoEnvWithPrefix(t *testing.T) {
	v := New()
	v.AutomaticEnv()
	v.SetEnvPrefix("Baz")
	os.Setenv("BAZ_BAR", "13")
	assert.Equal(t, "13", v.Get("bar"))
}

func TestSetEnvKeyReplacer(t *testing.T) {
	v := New()
	v.AutomaticEnv()
	os.Setenv("REFRESH_INTERVAL", "30s")

	replacer := strings.NewReplacer("-", "_")
	v.SetEnvKeyReplacer(replacer)

	assert.Equal(t, "30s", v.Get("refresh-interval"))
}

func TestAllKeys(t *testing.T) {
	v := New()
	initConfigs(v)

	ks := sort.StringSlice{"title", "newkey", "owner.organization", "owner.dob", "owner.bio", "name", "beard", "ppu", "batters.batter", "hobbies", "clothing.jacket", "clothing.trousers", "clothing.pants.size", "age", "hacker", "id", "type", "eyes", "p_id", "p_ppu", "p_batters.batter.type", "p_type", "p_name", "foos"}
	dob, _ := time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	all := map[string]interface{}{"owner": map[string]interface{}{"organization": "MongoDB", "bio": "MongoDB Chief Developer Advocate & Hacker at Large", "dob": dob}, "title": "TOML Example", "ppu": 0.55, "eyes": "brown", "clothing": map[string]interface{}{"trousers": "denim", "jacket": "leather", "pants": map[string]interface{}{"size": "large"}}, "id": "0001", "batters": map[string]interface{}{"batter": []interface{}{map[string]interface{}{"type": "Regular"}, map[string]interface{}{"type": "Chocolate"}, map[string]interface{}{"type": "Blueberry"}, map[string]interface{}{"type": "Devil's Food"}}}, "hacker": true, "beard": true, "hobbies": []interface{}{"skateboarding", "snowboarding", "go"}, "age": 35, "type": "donut", "newkey": "remote", "name": "Cake", "p_id": "0001", "p_ppu": "0.55", "p_name": "Cake", "p_batters": map[string]interface{}{"batter": map[string]interface{}{"type": "Regular"}}, "p_type": "donut", "foos": []map[string]interface{}{map[string]interface{}{"foo": []map[string]interface{}{map[string]interface{}{"key": 1}, map[string]interface{}{"key": 2}, map[string]interface{}{"key": 3}, map[string]interface{}{"key": 4}}}}}

	var allkeys sort.StringSlice
	allkeys = v.AllKeys()
	allkeys.Sort()
	ks.Sort()

	assert.Equal(t, ks, allkeys)
	assert.Equal(t, all, v.AllSettings())
}

func TestAllKeysWithEnv(t *testing.T) {
	v := New()

	// bind and define environment variables (including a nested one)
	v.BindEnv("id")
	v.BindEnv("foo.bar")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	os.Setenv("ID", "13")
	os.Setenv("FOO_BAR", "baz")

	expectedKeys := sort.StringSlice{"id", "foo.bar"}
	expectedKeys.Sort()
	keys := sort.StringSlice(v.AllKeys())
	keys.Sort()
	assert.Equal(t, expectedKeys, keys)
}

func TestAllSettingsWithDelimiterInKeys(t *testing.T) {
	exampleYaml := `Hacker: true
name: steve
hobbies:
- skate.boarding
- snowboarding
clothing.summer:
  jacket: leather
  pants:
    size: large
clothing.winter:
  jacket: wool
  pants:
    size: medium
`
	v := New()
	v.SetConfigType("yaml")
	r := strings.NewReader(exampleYaml)

	if err := v.unmarshalReader(r, v.config); err != nil {
		panic(err)
	}

	expectedAll := map[string]interface{}{"hacker": true, "name": "steve", "hobbies": []interface{}{"skate.boarding", "snowboarding"}, "clothing.summer": map[string]interface{}{"jacket": "leather", "pants": map[string]interface{}{"size": "large"}}, "clothing.winter": map[string]interface{}{"jacket": "wool", "pants": map[string]interface{}{"size": "medium"}}}
	expectedKeys := sort.StringSlice{"hacker", "name", "hobbies", "clothing.summer.jacket", "clothing.summer.pants.size", "clothing.winter.jacket", "clothing.winter.pants.size"}
	expectedKeys.Sort()
	keys := sort.StringSlice(v.AllKeys())
	keys.Sort()

	assert.Equal(t, expectedKeys, keys)
	assert.Equal(t, expectedAll, v.AllSettings())
}

func TestAliasesOfAliases(t *testing.T) {
	v := New()
	v.Set("Title", "Checking Case")
	v.RegisterAlias("Foo", "Bar")
	v.RegisterAlias("Bar", "Title")
	assert.Equal(t, "Checking Case", v.Get("FOO"))
}

func TestRecursiveAliases(t *testing.T) {
	v := New()
	v.RegisterAlias("Baz", "Roo")
	v.RegisterAlias("Roo", "baz")
}

func TestUnmarshal(t *testing.T) {
	v := New()
	v.SetDefault("port", 1313)
	v.Set("name", "Steve")
	v.Set("duration", "1s1ms")

	type config struct {
		Port     int
		Name     string
		Duration time.Duration
	}

	var C config

	err := v.Unmarshal(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	assert.Equal(t, &config{Name: "Steve", Port: 1313, Duration: time.Second + time.Millisecond}, &C)

	v.Set("port", 1234)
	err = v.Unmarshal(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}
	assert.Equal(t, &config{Name: "Steve", Port: 1234, Duration: time.Second + time.Millisecond}, &C)
}

func TestUnmarshalWithDecoderOptions(t *testing.T) {
	v := New()
	v.Set("credentials", "{\"foo\":\"bar\"}")

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

	err := v.Unmarshal(&C, opt)
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

	var testValues = map[string]*string{
		"host":     nil,
		"port":     nil,
		"endpoint": nil,
	}

	var mutatedTestValues = map[string]string{
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

func TestBindPFlagsStringSlice(t *testing.T) {
	tests := []struct {
		Expected []string
		Value    string
	}{
		{nil, ""},
		{[]string{"jeden"}, "jeden"},
		{[]string{"dwa", "trzy"}, "dwa,trzy"},
		{[]string{"cztery", "piec , szesc"}, "cztery,\"piec , szesc\""},
	}

	v := New()
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
			} else {
				assert.Equal(t, defaultVal, val.StringSlice)
			}
		}
	}
}

func TestBindPFlag(t *testing.T) {
	v := New()

	var testString = "testing"
	var testValue = newStringValue(testString, &testString)

	flag := &pflag.Flag{
		Name:    "testflag",
		Value:   testValue,
		Changed: false,
	}

	v.BindPFlag("testvalue", flag)

	assert.Equal(t, testString, v.Get("testvalue"))

	flag.Value.Set("testing_mutate")
	flag.Changed = true //hack for pflag usage

	assert.Equal(t, "testing_mutate", v.Get("testvalue"))

}

func TestBoundCaseSensitivity(t *testing.T) {
	v := New()
	initYAML(v)

	assert.Equal(t, "brown", v.Get("eyes"))

	v.BindEnv("eYEs", "TURTLE_EYES")
	os.Setenv("TURTLE_EYES", "blue")

	assert.Equal(t, "blue", v.Get("eyes"))

	var testString = "green"
	var testValue = newStringValue(testString, &testString)

	flag := &pflag.Flag{
		Name:    "eyeballs",
		Value:   testValue,
		Changed: true,
	}

	v.BindPFlag("eYEs", flag)
	assert.Equal(t, "green", v.Get("eyes"))

}

func TestSizeInBytes(t *testing.T) {
	input := map[string]struct {
		Size  uint
		Error bool
	}{
		"":               {0, true},
		"b":              {0, true},
		"12 bytes":       {0, true},
		"200000000000gb": {0, true},
		"12 b":           {12, false},
		"43 MB":          {43 * (1 << 20), false},
		"10mb":           {10 * (1 << 20), false},
		"1gb":            {1 << 30, false},
	}

	for str, expected := range input {
		size, err := parseSizeInBytes(str)
		assert.Equal(t, expected.Size, size, str)
		assert.Equal(t, expected.Error, err != nil, str)
	}
}

func TestFindsNestedKeys(t *testing.T) {
	v := New()
	initConfigs(v)
	dob, _ := time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")

	v.Set("super", map[string]interface{}{
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
		"title":  "TOML Example",
		"newkey": "remote",
		"batters": map[string]interface{}{
			"batter": []interface{}{
				map[string]interface{}{
					"type": "Regular",
				},
				map[string]interface{}{
					"type": "Chocolate",
				}, map[string]interface{}{
					"type": "Blueberry",
				}, map[string]interface{}{
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
			map[string]interface{}{
				"foo": []map[string]interface{}{
					map[string]interface{}{
						"key": 1,
					},
					map[string]interface{}{
						"key": 2,
					},
					map[string]interface{}{
						"key": 3,
					},
					map[string]interface{}{
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
	assert.False(t, v.InConfig("state"))
	assert.Equal(t, "steve", v.Get("name"))
	assert.Equal(t, []interface{}{"skateboarding", "snowboarding", "go"}, v.Get("hobbies"))
	assert.Equal(t, map[string]interface{}{"jacket": "leather", "trousers": "denim", "pants": map[string]interface{}{"size": "large"}}, v.Get("clothing"))
	assert.Equal(t, 35, v.Get("age"))
}

func TestIsSet(t *testing.T) {
	v := New()
	v.SetConfigType("yaml")
	v.ReadConfig(bytes.NewBuffer(yamlExample))
	assert.True(t, v.IsSet("clothing.jacket"))
	assert.False(t, v.IsSet("clothing.jackets"))
	assert.False(t, v.IsSet("helloworld"))
	v.Set("helloworld", "fubar")
	assert.True(t, v.IsSet("helloworld"))
}

func TestDirsSearch(t *testing.T) {

	root, config, cleanup := initDirs(t)
	defer cleanup()

	v := New()
	v.SetConfigName(config)
	v.SetDefault(`key`, `default`)

	entries, err := ioutil.ReadDir(root)
	for _, e := range entries {
		if e.IsDir() {
			v.AddConfigPath(e.Name())
		}
	}

	err = v.ReadInConfig()
	assert.Nil(t, err)

	assert.Equal(t, `value is `+path.Base(v.configPaths[0]), v.GetString(`key`))
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

func TestWriteConfigHCL(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	v.SetFs(fs)
	v.SetConfigName("c")
	v.SetConfigType("hcl")
	err := v.ReadConfig(bytes.NewBuffer(hclExample))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.WriteConfigAs("c.hcl"); err != nil {
		t.Fatal(err)
	}
	read, err := afero.ReadFile(fs, "c.hcl")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, hclWriteExpected, read)
}

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

func TestWriteConfigJson(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	v.SetFs(fs)
	v.SetConfigName("c")
	v.SetConfigType("json")
	err := v.ReadConfig(bytes.NewBuffer(jsonExample))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.WriteConfigAs("c.json"); err != nil {
		t.Fatal(err)
	}
	read, err := afero.ReadFile(fs, "c.json")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, jsonWriteExpected, read)
}

var propertiesWriteExpected = []byte(`p_id = 0001
p_type = donut
p_name = Cake
p_ppu = 0.55
p_batters.batter.type = Regular
`)

func TestWriteConfigProperties(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	v.SetFs(fs)
	v.SetConfigName("c")
	v.SetConfigType("properties")
	err := v.ReadConfig(bytes.NewBuffer(propertiesExample))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.WriteConfigAs("c.properties"); err != nil {
		t.Fatal(err)
	}
	read, err := afero.ReadFile(fs, "c.properties")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, propertiesWriteExpected, read)
}

func TestWriteConfigTOML(t *testing.T) {
	fs := afero.NewMemMapFs()
	v := New()
	v.SetFs(fs)
	v.SetConfigName("c")
	v.SetConfigType("toml")
	err := v.ReadConfig(bytes.NewBuffer(tomlExample))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.WriteConfigAs("c.toml"); err != nil {
		t.Fatal(err)
	}

	// The TOML String method does not order the contents.
	// Therefore, we must read the generated file and compare the data.
	v2 := New()
	v2.SetFs(fs)
	v2.SetConfigName("c")
	v2.SetConfigType("toml")
	v2.SetConfigFile("c.toml")
	err = v2.ReadInConfig()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, v.GetString("title"), v2.GetString("title"))
	assert.Equal(t, v.GetString("owner.bio"), v2.GetString("owner.bio"))
	assert.Equal(t, v.GetString("owner.dob"), v2.GetString("owner.dob"))
	assert.Equal(t, v.GetString("owner.organization"), v2.GetString("owner.organization"))
}

var yamlWriteExpected = []byte(`age: 35
beard: true
clothing:
  jacket: leather
  pants:
    size: large
  trousers: denim
eyes: brown
hacker: true
hobbies:
- skateboarding
- snowboarding
- go
name: steve
`)

func TestWriteConfigYAML(t *testing.T) {
	v := New()
	fs := afero.NewMemMapFs()
	v.SetFs(fs)
	v.SetConfigName("c")
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(yamlExample))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.WriteConfigAs("c.yaml"); err != nil {
		t.Fatal(err)
	}
	read, err := afero.ReadFile(fs, "c.yaml")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, yamlWriteExpected, read)
}

var yamlMergeExampleTgt = []byte(`
hello:
    pop: 37890
    lagrenum: 765432101234567
    world:
    - us
    - uk
    - fr
    - de
`)

var yamlMergeExampleSrc = []byte(`
hello:
    pop: 45000
    lagrenum: 7654321001234567
    universe:
    - mw
    - ad
fu: bar
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

	if pop := v.GetInt64("hello.lagrenum"); pop != int64(765432101234567) {
		t.Fatalf("int64 lagrenum != 765432101234567, = %d", pop)
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

	if pop := v.GetInt64("hello.lagrenum"); pop != int64(7654321001234567) {
		t.Fatalf("int64 lagrenum != 7654321001234567, = %d", pop)
	}

	if world := v.GetStringSlice("hello.world"); len(world) != 4 {
		t.Fatalf("len(world) != 4, = %d", len(world))
	}

	if universe := v.GetStringSlice("hello.universe"); len(universe) != 2 {
		t.Fatalf("len(universe) != 2, = %d", len(universe))
	}

	if fu := v.GetString("fu"); fu != "bar" {
		t.Fatalf("fu != \"bar\", = %s", fu)
	}
}

func TestMergeConfigOverride(t *testing.T) {
	v := New()
	v.SetConfigType("yml")

	v.SetEnvPrefix("Baz")
	v.BindEnv("fu")
	os.Setenv("BAZ_FU", "test")

	if err := v.ReadConfig(bytes.NewBuffer(yamlMergeExampleTgt)); err != nil {
		t.Fatal(err)
	}

	if pop := v.GetInt("hello.pop"); pop != 37890 {
		t.Fatalf("pop != 37890, = %d", pop)
	}

	if pop := v.GetInt("hello.lagrenum"); pop != 765432101234567 {
		t.Fatalf("lagrenum != 765432101234567, = %d", pop)
	}

	if pop := v.GetInt32("hello.pop"); pop != int32(37890) {
		t.Fatalf("pop != 37890, = %d", pop)
	}

	if pop := v.GetInt64("hello.lagrenum"); pop != int64(765432101234567) {
		t.Fatalf("int64 lagrenum != 765432101234567, = %d", pop)
	}

	if world := v.GetStringSlice("hello.world"); len(world) != 4 {
		t.Fatalf("len(world) != 4, = %d", len(world))
	}

	if fu := v.GetString("fu"); fu != "test" {
		t.Fatalf("fu != \"test\", = %s", fu)
	}

	if err := v.MergeConfigOverride(bytes.NewBuffer(yamlMergeExampleSrc)); err != nil {
		t.Fatal(err)
	}

	if pop := v.GetInt("hello.pop"); pop != 45000 {
		t.Fatalf("pop != 45000, = %d", pop)
	}

	if pop := v.GetInt("hello.lagrenum"); pop != 7654321001234567 {
		t.Fatalf("lagrenum != 7654321001234567, = %d", pop)
	}

	if pop := v.GetInt32("hello.pop"); pop != int32(45000) {
		t.Fatalf("pop != 45000, = %d", pop)
	}

	if pop := v.GetInt64("hello.lagrenum"); pop != int64(7654321001234567) {
		t.Fatalf("int64 lagrenum != 7654321001234567, = %d", pop)
	}

	if world := v.GetStringSlice("hello.world"); len(world) != 4 {
		t.Fatalf("len(world) != 4, = %d", len(world))
	}

	if universe := v.GetStringSlice("hello.universe"); len(universe) != 2 {
		t.Fatalf("len(universe) != 2, = %d", len(universe))
	}

	if fu := v.GetString("fu"); fu != "bar" {
		t.Fatalf("fu != \"bar\", = %s", fu)
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
		large := v.GetInt("hello.lagrenum")
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
	v := New()
	v.SetConfigFile("/tmp/config.yaml")
	v.SetConfigName("default")
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
	v := New()
	initConfig(v, "yaml", config)

	assert.Equal(t, "steve", v.GetString("name"))

	polyester := "polyester"
	v.SetDefault("clothing.shirt", polyester)
	v.SetDefault("clothing.jacket.price", 100)

	assert.Equal(t, "leather", v.GetString("clothing.jacket"))
	assert.Nil(t, v.Get("clothing.jacket.price"))
	assert.Equal(t, polyester, v.GetString("clothing.shirt"))

	clothingSettings := v.AllSettings()["clothing"].(map[string]interface{})
	assert.Equal(t, "leather", clothingSettings["jacket"])
	assert.Equal(t, polyester, clothingSettings["shirt"])
}

func TestDotParameter(t *testing.T) {
	v := New()
	initJSON(v)
	// shoud take precedence over batters defined in jsonExample
	r := bytes.NewReader([]byte(`{ "batters.batter": [ { "type": "Small" } ] }`))
	v.unmarshalReader(r, v.config)

	actual := v.Get("batters.batter")
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
	v := New()
	m1 := map[string]interface{}{
		"Foo": 32,
		"Bar": map[interface{}]interface {
		}{
			"ABc": "A",
			"cDE": "B"},
	}

	m2 := map[string]interface{}{
		"Foo": 52,
		"Bar": map[interface{}]interface {
		}{
			"bCd": "A",
			"eFG": "B"},
	}

	v.Set("Given1", m1)
	v.Set("Number1", 42)

	v.SetDefault("Given2", m2)
	v.SetDefault("Number2", 52)

	// Verify SetDefault
	if v := v.Get("number2"); v != 52 {
		t.Fatalf("Expected 52 got %q", v)
	}

	if v := v.Get("given2.foo"); v != 52 {
		t.Fatalf("Expected 52 got %q", v)
	}

	if v := v.Get("given2.bar.bcd"); v != "A" {
		t.Fatalf("Expected A got %q", v)
	}

	if _, ok := m2["Foo"]; !ok {
		t.Fatal("Input map changed")
	}

	// Verify Set
	if v := v.Get("number1"); v != 42 {
		t.Fatalf("Expected 42 got %q", v)
	}

	if v := v.Get("given1.foo"); v != 32 {
		t.Fatalf("Expected 32 got %q", v)
	}

	if v := v.Get("given1.bar.abc"); v != "A" {
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
	v := New()
	initConfig(v, "toml", config)

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
	v := New()
	initConfig(v, typ, config)
	v.Set("RfD", true)
	assert.Equal(t, true, v.Get("rfd"))
	assert.Equal(t, true, v.Get("rFD"))
	assert.Equal(t, 1, cast.ToInt(v.Get("abcd")))
	assert.Equal(t, 1, cast.ToInt(v.Get("Abcd")))
	assert.Equal(t, 2, cast.ToInt(v.Get("ef.gh")))
	assert.Equal(t, 3, cast.ToInt(v.Get("ef.ijk")))
	assert.Equal(t, 4, cast.ToInt(v.Get("ef.lm.no")))
	assert.Equal(t, 5, cast.ToInt(v.Get("ef.lm.p.q")))

}

func newViperWithConfigFile(t *testing.T) (*Viper, string, func()) {
	watchDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	configFile := path.Join(watchDir, "config.yaml")
	err = ioutil.WriteFile(configFile, []byte("foo: bar\n"), 0640)
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
	err = os.Mkdir(dataDir1, 0777)
	require.Nil(t, err)
	realConfigFile := path.Join(dataDir1, "config.yaml")
	t.Logf("Real config file location: %s\n", realConfigFile)
	err = ioutil.WriteFile(realConfigFile, []byte("foo: bar\n"), 0640)
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
		v.OnConfigChange(func(in fsnotify.Event) {
			t.Logf("config file changed")
			wg.Done()
		})
		v.WatchConfig()
		// when overwriting the file and waiting for the custom change notification handler to be triggered
		err = ioutil.WriteFile(configFile, []byte("foo: baz\n"), 0640)
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
		err := os.Mkdir(dataDir2, 0777)
		require.Nil(t, err)
		configFile2 := path.Join(dataDir2, "config.yaml")
		err = ioutil.WriteFile(configFile2, []byte("foo: baz\n"), 0640)
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

func BenchmarkGetBool(b *testing.B) {
	key := "BenchmarkGetBool"
	v := New()
	v.Set(key, true)

	for i := 0; i < b.N; i++ {
		if !v.GetBool(key) {
			b.Fatal("GetBool returned false")
		}
	}
}

func BenchmarkGet(b *testing.B) {
	key := "BenchmarkGet"
	v := New()
	v.Set(key, true)

	for i := 0; i < b.N; i++ {
		if !v.Get(key).(bool) {
			b.Fatal("Get returned false")
		}
	}
}

// This is the "perfect result" for the above.
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

func TestKnownKeys(t *testing.T) {
	v := New()
	v.SetDefault("default", 45)
	if _, ok := v.GetKnownKeys()["default"]; !ok {
		t.Error("SetDefault didn't mark key as known")
	}
	v.BindEnv("bind", "my_env_var")
	if _, ok := v.GetKnownKeys()["bind"]; !ok {
		t.Error("BindEnv didn't mark key as known")
	}
	v.RegisterAlias("my_alias", "key")
	if _, ok := v.GetKnownKeys()["my_alias"]; !ok {
		t.Error("RegisterAlias didn't mark alias as known")
	}
	v.SetKnown("known")
	if _, ok := v.GetKnownKeys()["known"]; !ok {
		t.Error("SetKnown didn't mark key as known")
	}
}

func TestEmptySection(t *testing.T) {

	var yamlWithEnvVars = `
key:
    subkey:
another_key:
`

	v := New()
	initConfig(v, "yaml", yamlWithEnvVars)
	v.SetKnown("is_known")
	v.SetDefault("has_default", true)

	// AllKeys includes empty keys
	keys := v.AllKeys()
	assert.NotContains(t, keys, "key") // Only empty leaf nodes are returned
	assert.Contains(t, keys, "key.subkey")
	assert.Contains(t, keys, "another_key")
	assert.NotContains(t, keys, "is_known")
	assert.Contains(t, keys, "has_default")

	// AllSettings includes empty keys
	vars := v.AllSettings()
	assert.NotContains(t, vars, "key") // Only empty leaf nodes are returned
	assert.Contains(t, vars, "key.subkey")
	assert.Contains(t, vars, "another_key")
	assert.NotContains(t, vars, "is_known")
	assert.Contains(t, vars, "has_default")
}
