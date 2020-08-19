package viper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/magiconair/properties"
	"github.com/pelletier/go-toml"
	"github.com/spf13/afero"
	"github.com/subosito/gotenv"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

var SupportedParsers map[string]Parser

type Parser interface {
	UnmarshalReader(v *Viper, in io.Reader, c map[string]interface{}) error // Unmarshal a Reader into a map.
	MarshalWriter(v *Viper, f afero.File, c map[string]interface{}) error   // Marshal a map into Writer.
}

// add support parser
func AddParser(parser Parser, names ...string) {
	for _, n := range names {
		SupportedParsers[n] = parser
	}
}

func parserInit() {
	SupportedParsers = make(map[string]Parser, 32)
	AddParser(&JSONParser{}, "json")
	AddParser(&TOMLParser{}, "toml")
	AddParser(&YAMLParser{}, "yaml", "yml")
	AddParser(&PROPSParser{}, "properties", "props", "prop")
	AddParser(&HCLParser{}, "hcl")
	AddParser(&DOTENVParser{}, "dotenv", "env")
	AddParser(&INIParser{}, "ini")
}

func parserReset() {
	parserInit()
}

// json parser
type JSONParser struct {
}

func (pp *JSONParser) UnmarshalReader(v *Viper, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	return json.Unmarshal(buf.Bytes(), &c)
}
func (pp *JSONParser) MarshalWriter(v *Viper, f afero.File, c map[string]interface{}) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.WriteString(string(b))
	return err
}

// toml parser
type TOMLParser struct {
}

func (pp *TOMLParser) UnmarshalReader(v *Viper, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	tree, err := toml.LoadReader(buf)
	if err != nil {
		return err
	}
	tmap := tree.ToMap()
	for k, v := range tmap {
		c[k] = v
	}
	return nil
}
func (pp *TOMLParser) MarshalWriter(v *Viper, f afero.File, c map[string]interface{}) error {
	t, err := toml.TreeFromMap(c)
	if err != nil {
		return err
	}
	s := t.String()
	if _, err := f.WriteString(s); err != nil {
		return err
	}
	return nil
}

// yaml parser
type YAMLParser struct {
}

func (pp *YAMLParser) UnmarshalReader(v *Viper, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	return yaml.Unmarshal(buf.Bytes(), &c)
}
func (pp *YAMLParser) MarshalWriter(v *Viper, f afero.File, c map[string]interface{}) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if _, err = f.WriteString(string(b)); err != nil {
		return err
	}
	return nil
}

// ini parser
type INIParser struct {
}

func (pp *INIParser) UnmarshalReader(v *Viper, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	cfg := ini.Empty()
	err := cfg.Append(buf.Bytes())
	if err != nil {
		return err
	}
	sections := cfg.Sections()
	for i := 0; i < len(sections); i++ {
		section := sections[i]
		keys := section.Keys()
		for j := 0; j < len(keys); j++ {
			key := keys[j]
			value := cfg.Section(section.Name()).Key(key.Name()).String()
			c[section.Name()+"."+key.Name()] = value
		}
	}
	return nil
}
func (pp *INIParser) MarshalWriter(v *Viper, f afero.File, c map[string]interface{}) error {
	keys := v.AllKeys()
	cfg := ini.Empty()
	ini.PrettyFormat = false
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		lastSep := strings.LastIndex(key, ".")
		sectionName := key[:(lastSep)]
		keyName := key[(lastSep + 1):]
		if sectionName == "default" {
			sectionName = ""
		}
		cfg.Section(sectionName).Key(keyName).SetValue(v.Get(key).(string))
	}
	cfg.WriteTo(f)
	return nil
}

// hcl parser
type HCLParser struct {
}

func (pp *HCLParser) UnmarshalReader(v *Viper, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	obj, err := hcl.Parse(buf.String())
	if err != nil {
		return err
	}
	if err = hcl.DecodeObject(&c, obj); err != nil {
		return err
	}
	return nil
}
func (pp *HCLParser) MarshalWriter(v *Viper, f afero.File, c map[string]interface{}) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	ast, err := hcl.Parse(string(b))
	if err != nil {
		return err
	}
	err = printer.Fprint(f, ast.Node)
	if err != nil {
		return err
	}
	return nil
}

// dot env parser
type DOTENVParser struct {
}

func (pp *DOTENVParser) UnmarshalReader(v *Viper, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	env, err := gotenv.StrictParse(buf)
	if err != nil {
		return err
	}
	for k, v := range env {
		c[k] = v
	}

	return nil
}
func (pp *DOTENVParser) MarshalWriter(v *Viper, f afero.File, c map[string]interface{}) error {
	lines := []string{}
	for _, key := range v.AllKeys() {
		envName := strings.ToUpper(strings.Replace(key, ".", "_", -1))
		val := v.Get(key)
		lines = append(lines, fmt.Sprintf("%v=%v", envName, val))
	}
	s := strings.Join(lines, "\n")
	if _, err := f.WriteString(s); err != nil {
		return err
	}
	return nil
}

// props parser
type PROPSParser struct {
}

func (pp *PROPSParser) UnmarshalReader(v *Viper, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	v.properties = properties.NewProperties()
	var err error
	if v.properties, err = properties.Load(buf.Bytes(), properties.UTF8); err != nil {
		return ConfigParseError{err}
	}
	for _, key := range v.properties.Keys() {
		value, _ := v.properties.Get(key)
		// recursively build nested maps
		path := strings.Split(key, ".")
		lastKey := strings.ToLower(path[len(path)-1])
		deepestMap := deepSearch(c, path[0:len(path)-1])
		// set innermost value
		deepestMap[lastKey] = value
	}
	return nil
}
func (pp *PROPSParser) MarshalWriter(v *Viper, f afero.File, c map[string]interface{}) error {
	if v.properties == nil {
		v.properties = properties.NewProperties()
	}
	p := v.properties
	for _, key := range v.AllKeys() {
		_, _, err := p.Set(key, v.GetString(key))
		if err != nil {
			return err
		}
	}
	_, err := p.WriteComment(f, "#", properties.UTF8)
	if err != nil {
		return err
	}
	return nil
}
