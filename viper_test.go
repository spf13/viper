// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package viper

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
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
age: 35
eyes : brown
beard: true
`)

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

var remoteExample = []byte(`{
"id":"0002",
"type":"cronut",
"newkey":"remote"
}`)

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
	SetConfigFile("/tmp/config.yaml")
	assert.Equal(t, "/tmp/config.yaml", getConfigFile())
}

func TestDefault(t *testing.T) {
	SetDefault("age", 45)
	assert.Equal(t, 45, Get("age"))
}

func TestMarshalling(t *testing.T) {
	SetConfigType("yaml")
	r := bytes.NewReader(yamlExample)

	MarshallReader(r, config)
	assert.True(t, InConfig("name"))
	assert.False(t, InConfig("state"))
	assert.Equal(t, "steve", Get("name"))
	assert.Equal(t, []interface{}{"skateboarding", "snowboarding", "go"}, Get("hobbies"))
	assert.Equal(t, map[interface{}]interface{}{"jacket": "leather", "trousers": "denim"}, Get("clothing"))
	assert.Equal(t, 35, Get("age"))
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
	Reset()
	SetConfigType("yml")
	r := bytes.NewReader(yamlExample)

	MarshallReader(r, config)
	assert.Equal(t, "steve", Get("name"))
}

func TestJSON(t *testing.T) {
	SetConfigType("json")
	r := bytes.NewReader(jsonExample)

	MarshallReader(r, config)
	assert.Equal(t, "0001", Get("id"))
}

func TestTOML(t *testing.T) {
	SetConfigType("toml")
	r := bytes.NewReader(tomlExample)

	MarshallReader(r, config)
	assert.Equal(t, "TOML Example", Get("title"))
}

func TestRemotePrecedence(t *testing.T) {
	SetConfigType("json")
	r := bytes.NewReader(jsonExample)
	MarshallReader(r, config)
	remote := bytes.NewReader(remoteExample)
	assert.Equal(t, "0001", Get("id"))
	MarshallReader(remote, kvstore)
	assert.Equal(t, "0001", Get("id"))
	assert.NotEqual(t, "cronut", Get("type"))
	assert.Equal(t, "remote", Get("newkey"))
	Set("newkey", "newvalue")
	assert.NotEqual(t, "remote", Get("newkey"))
	assert.Equal(t, "newvalue", Get("newkey"))
	Set("newkey", "remote")
}

func TestEnv(t *testing.T) {
	SetConfigType("json")
	r := bytes.NewReader(jsonExample)
	MarshallReader(r, config)
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

func TestAllKeys(t *testing.T) {
	ks := sort.StringSlice{"title", "newkey", "owner", "name", "beard", "ppu", "batters", "hobbies", "clothing", "age", "hacker", "id", "type", "eyes"}
	dob, _ := time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	all := map[string]interface{}{"hacker": true, "beard": true, "newkey": "remote", "batters": map[string]interface{}{"batter": []interface{}{map[string]interface{}{"type": "Regular"}, map[string]interface{}{"type": "Chocolate"}, map[string]interface{}{"type": "Blueberry"}, map[string]interface{}{"type": "Devil's Food"}}}, "hobbies": []interface{}{"skateboarding", "snowboarding", "go"}, "ppu": 0.55, "clothing": map[interface{}]interface{}{"jacket": "leather", "trousers": "denim"}, "name": "crunk", "owner": map[string]interface{}{"organization": "MongoDB", "Bio": "MongoDB Chief Developer Advocate & Hacker at Large", "dob": dob}, "id": "13", "title": "TOML Example", "age": 35, "type": "donut", "eyes": "brown"}

	var allkeys sort.StringSlice
	allkeys = AllKeys()
	allkeys.Sort()
	ks.Sort()

	assert.Equal(t, ks, allkeys)
	assert.Equal(t, all, AllSettings())
}

func TestCaseInSensitive(t *testing.T) {
	assert.Equal(t, true, Get("hacker"))
	Set("Title", "Checking Case")
	assert.Equal(t, "Checking Case", Get("tItle"))
}

func TestAliasesOfAliases(t *testing.T) {
	RegisterAlias("Foo", "Bar")
	RegisterAlias("Bar", "Title")
	assert.Equal(t, "Checking Case", Get("FOO"))
}

func TestRecursiveAliases(t *testing.T) {
	RegisterAlias("Baz", "Roo")
	RegisterAlias("Roo", "baz")
}

func TestMarshal(t *testing.T) {
	SetDefault("port", 1313)
	Set("name", "Steve")

	type config struct {
		Port int
		Name string
	}

	var C config

	err := Marshal(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	assert.Equal(t, &C, &config{Name: "Steve", Port: 1313})

	Set("port", 1234)
	err = Marshal(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}
	assert.Equal(t, &C, &config{Name: "Steve", Port: 1234})
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
