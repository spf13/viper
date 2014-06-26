// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package viper

import (
	"bytes"
	"testing"

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
age: 35`)

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

	MarshallReader(r)
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

func TestYML(t *testing.T) {
	Reset()
	SetConfigType("yml")
	r := bytes.NewReader(yamlExample)

	MarshallReader(r)
	assert.Equal(t, "steve", Get("name"))
}

func TestJSON(t *testing.T) {
	SetConfigType("json")
	r := bytes.NewReader(jsonExample)

	MarshallReader(r)
	assert.Equal(t, "0001", Get("id"))
}

func TestTOML(t *testing.T) {
	SetConfigType("toml")
	r := bytes.NewReader(tomlExample)

	MarshallReader(r)
	assert.Equal(t, "TOML Example", Get("title"))
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

func TestIntoStruct(t *testing.T) {
	SetDefault("port", 1313)
	Set("name", "Steve")

	type config struct {
		Port int
		Name string
	}

	var C config

	err := GetAllIntoStruct(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}

	assert.Equal(t, &C, &config{Name: "Steve", Port: 1313})

	Set("port", 1234)
	err = GetAllIntoStruct(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}
	assert.Equal(t, &C, &config{Name: "Steve", Port: 1234})
}
