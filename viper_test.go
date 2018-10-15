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
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"github.com/spf13/cast"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

type testUnmarshalExtra struct {
	Existing bool
}

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

func initConfigs() {
	Reset()
	var r io.Reader

	SetConfigType("json")
	r = bytes.NewReader(jsonExample)
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

func initJSON() {
	Reset()
	SetConfigType("json")
	r := bytes.NewReader(jsonExample)

	unmarshalReader(r, v.config)
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

func TestJSON(t *testing.T) {
	initJSON()
	assert.Equal(t, "0001", Get("id"))
}

func TestEnv(t *testing.T) {
	initJSON()

	BindEnv("id")
	BindEnv("f", "FOOD")

	os.Setenv("ID", "13")
	os.Setenv("FOOD", "apple")
	os.Setenv("NAME", "crunk")

	assert.Equal(t, "13", Get("id"))
	assert.Equal(t, "apple", Get("f"))
	assert.Equal(t, "Cake", Get("name"))

	AutomaticEnv()

	assert.Equal(t, "crunk", Get("name"))

}

func TestEnvPrefix(t *testing.T) {
	initJSON()

	SetEnvPrefix("foo") // will be uppercased automatically
	BindEnv("id")
	BindEnv("f", "FOOD") // not using prefix

	os.Setenv("FOO_ID", "13")
	os.Setenv("FOOD", "apple")
	os.Setenv("FOO_NAME", "crunk")

	assert.Equal(t, "13", Get("id"))
	assert.Equal(t, "apple", Get("f"))
	assert.Equal(t, "Cake", Get("name"))

	AutomaticEnv()

	assert.Equal(t, "crunk", Get("name"))
}

func TestAutoEnv(t *testing.T) {
	Reset()

	AutomaticEnv()
	os.Setenv("FOO_BAR", "13")
	assert.Equal(t, "13", Get("foo_bar"))
}

func TestAutoEnvWithPrefix(t *testing.T) {
	Reset()

	AutomaticEnv()
	SetEnvPrefix("Baz")
	os.Setenv("BAZ_BAR", "13")
	assert.Equal(t, "13", Get("bar"))
}

func TestSetEnvKeyReplacer(t *testing.T) {
	Reset()

	AutomaticEnv()
	os.Setenv("REFRESH_INTERVAL", "30s")

	replacer := strings.NewReplacer("-", "_")
	SetEnvKeyReplacer(replacer)

	assert.Equal(t, "30s", Get("refresh-interval"))
}

func TestAllKeys(t *testing.T) {
	initConfigs()

	ks := sort.StringSlice{"title", "newkey", "owner.organization", "owner.dob", "owner.bio", "name", "beard", "ppu", "batters.batter", "hobbies", "clothing.jacket", "clothing.trousers", "clothing.pants.size", "age", "hacker", "id", "type", "eyes", "p_id", "p_ppu", "p_batters.batter.type", "p_type", "p_name", "foos"}
	dob, _ := time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	all := map[string]interface{}{"owner": map[string]interface{}{"organization": "MongoDB", "bio": "MongoDB Chief Developer Advocate & Hacker at Large", "dob": dob}, "title": "TOML Example", "ppu": 0.55, "eyes": "brown", "clothing": map[string]interface{}{"trousers": "denim", "jacket": "leather", "pants": map[string]interface{}{"size": "large"}}, "id": "0001", "batters": map[string]interface{}{"batter": []interface{}{map[string]interface{}{"type": "Regular"}, map[string]interface{}{"type": "Chocolate"}, map[string]interface{}{"type": "Blueberry"}, map[string]interface{}{"type": "Devil's Food"}}}, "hacker": true, "beard": true, "hobbies": []interface{}{"skateboarding", "snowboarding", "go"}, "age": 35, "type": "donut", "newkey": "remote", "name": "Cake", "p_id": "0001", "p_ppu": "0.55", "p_name": "Cake", "p_batters": map[string]interface{}{"batter": map[string]interface{}{"type": "Regular"}}, "p_type": "donut", "foos": []map[string]interface{}{map[string]interface{}{"foo": []map[string]interface{}{map[string]interface{}{"key": 1}, map[string]interface{}{"key": 2}, map[string]interface{}{"key": 3}, map[string]interface{}{"key": 4}}}}}

	var allkeys sort.StringSlice
	allkeys = AllKeys()
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
	os.Setenv("ID", "13")
	os.Setenv("FOO_BAR", "baz")

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

	type config struct {
		Port     int
		Name     string
		Duration time.Duration
	}

	var C config

	err := Unmarshal(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	assert.Equal(t, &config{Name: "Steve", Port: 1313, Duration: time.Second + time.Millisecond}, &C)

	Set("port", 1234)
	err = Unmarshal(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}
	assert.Equal(t, &config{Name: "Steve", Port: 1234, Duration: time.Second + time.Millisecond}, &C)
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
	for _, testValue := range []struct {
		Expected []string
		Value    string
	}{
		{[]string{}, ""},
		{[]string{"jeden"}, "jeden"},
		{[]string{"dwa", "trzy"}, "dwa,trzy"},
		{[]string{"cztery", "piec , szesc"}, "cztery,\"piec , szesc\""}} {

		for _, changed := range []bool{true, false} {
			v := New() // create independent Viper object
			flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flagSet.StringSlice("stringslice", testValue.Expected, "test")
			flagSet.Visit(func(f *pflag.Flag) {
				if len(testValue.Value) > 0 {
					f.Value.Set(testValue.Value)
					f.Changed = changed
				}
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
			assert.Equal(t, testValue.Expected, val.StringSlice)
		}
	}
}

func TestBindPFlag(t *testing.T) {
	var testString = "testing"
	var testValue = newStringValue(testString, &testString)

	flag := &pflag.Flag{
		Name:    "testflag",
		Value:   testValue,
		Changed: false,
	}

	BindPFlag("testvalue", flag)

	assert.Equal(t, testString, Get("testvalue"))

	flag.Value.Set("testing_mutate")
	flag.Changed = true //hack for pflag usage

	assert.Equal(t, "testing_mutate", Get("testvalue"))

}

func TestBoundCaseSensitivity(t *testing.T) {
	assert.Equal(t, "brown", Get("eyes"))

	BindEnv("eYEs", "TURTLE_EYES")
	os.Setenv("TURTLE_EYES", "blue")

	assert.Equal(t, "blue", Get("eyes"))

	var testString = "green"
	var testValue = newStringValue(testString, &testString)

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

func TestCaseInsensitiveSet(t *testing.T) {
	Reset()
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
