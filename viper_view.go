package viper

import (
	"strings"
	"time"
)

type ViperView interface {
	Get(key string) interface{}
	Sub(key string) *Viper
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetFloat64(key string) float64
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetSizeInBytes(key string) uint
	UnmarshalKey(key string, rawVal interface{}, opts ...DecoderConfigOption) error
	Unmarshal(rawVal interface{}, opts ...DecoderConfigOption) error
	UnmarshalExact(rawVal interface{}) error
	IsSet(key string) bool
	InConfig(key string) bool
	AllKeys() []string
	AllSettings() map[string]interface{}
	GetConfigView(key string) ViperView
}

type viperView struct {
	viper        *Viper
	configPrefix string
}

func NewView() viperView {
	v := viperView{
		viper:        New(),
		configPrefix: "",
	}
	return v
}

func GetConfigView(key string) ViperView { return v.GetConfigView(key) }
func (v *Viper) GetConfigView(key string) ViperView {
	subv := NewView()
	subv.viper = v
	subv.configPrefix = key + subv.viper.keyDelim

	return &subv
}

func (v *viperView) GetConfigView(key string) ViperView {
	subv := NewView()
	subv.viper = v.viper
	subv.configPrefix = v.configPrefix + key + subv.viper.keyDelim

	return &subv
}

func (v *viperView) getKeyFullPath(key string) string {
	return v.configPrefix + key
}

func (v *viperView) Get(key string) interface{} {
	return v.viper.Get(v.configPrefix + key)
}

func (v *viperView) IsSet(key string) bool {
	return v.viper.IsSet(v.configPrefix + key)
}

func (v *viperView) Sub(key string) *Viper {
	return v.viper.Sub(v.getKeyFullPath(key))
}

func (v *viperView) GetString(key string) string {
	return v.viper.GetString(v.getKeyFullPath(key))
}

func (v *viperView) GetBool(key string) bool {
	return v.viper.GetBool(v.getKeyFullPath(key))
}

func (v *viperView) GetInt(key string) int {
	return v.viper.GetInt(v.getKeyFullPath(key))
}

func (v *viperView) GetInt32(key string) int32 {
	return v.viper.GetInt32(v.getKeyFullPath(key))
}

func (v *viperView) GetInt64(key string) int64 {
	return v.viper.GetInt64(v.getKeyFullPath(key))
}

func (v *viperView) GetFloat64(key string) float64 {
	return v.viper.GetFloat64(v.getKeyFullPath(key))
}

func (v *viperView) GetTime(key string) time.Time {
	return v.viper.GetTime(v.getKeyFullPath(key))
}

func (v *viperView) GetDuration(key string) time.Duration {
	return v.viper.GetDuration(v.getKeyFullPath(key))
}

func (v *viperView) GetStringSlice(key string) []string {
	return v.viper.GetStringSlice(v.getKeyFullPath(key))
}

func (v *viperView) GetStringMap(key string) map[string]interface{} {
	return v.viper.GetStringMap(v.getKeyFullPath(key))
}

func (v *viperView) GetStringMapString(key string) map[string]string {
	return v.viper.GetStringMapString(v.getKeyFullPath(key))
}

func (v *viperView) GetStringMapStringSlice(key string) map[string][]string {
	return v.viper.GetStringMapStringSlice(v.getKeyFullPath(key))
}

func (v *viperView) GetSizeInBytes(key string) uint {
	return v.viper.GetSizeInBytes(v.getKeyFullPath(key))
}

func (v *viperView) UnmarshalKey(key string, rawVal interface{}, opts ...DecoderConfigOption) error {
	return v.viper.UnmarshalKey(v.getKeyFullPath(key), rawVal, opts...)
}

func (v *viperView) InConfig(key string) bool {
	return v.viper.InConfig(v.getKeyFullPath(key))
}

func (v *viperView) AllSettings() map[string]interface{} {
	m := map[string]interface{}{}
	// start from the list of keys, and construct the map one value at a time
	for _, k := range v.AllKeys() {
		value := v.Get(k)
		if value == nil {
			// should not happen, since AllKeys() returns only keys holding a value,
			// check just in case anything changes
			continue
		}
		path := strings.Split(k, v.viper.keyDelim)
		lastKey := strings.ToLower(path[len(path)-1])
		deepestMap := deepSearch(m, path[0:len(path)-1])
		// set innermost value
		deepestMap[lastKey] = value
	}
	return m
}

func (v *viperView) AllKeys() []string {
	m := map[string]bool{}
	// add all paths, by order of descending priority to ensure correct shadowing
	m = v.viper.flattenAndMergeMap(m, castMapStringToMapInterface(v.viper.aliases), "")
	m = v.viper.flattenAndMergeMap(m, v.viper.override, "")
	m = v.viper.mergeFlatMap(m, castMapFlagToMapInterface(v.viper.pflags))
	m = v.viper.mergeFlatMap(m, castMapStringToMapInterface(v.viper.env))
	m = v.viper.flattenAndMergeMap(m, v.viper.config, "")
	m = v.viper.flattenAndMergeMap(m, v.viper.kvstore, "")
	m = v.viper.flattenAndMergeMap(m, v.viper.defaults, "")

	// convert set of paths to list
	a := []string{}
	for x := range m {
		if strings.HasPrefix(x, v.configPrefix) {
			a = append(a, strings.TrimPrefix(x, v.configPrefix))
		}
	}
	return a
}

func (v *viperView) Unmarshal(rawVal interface{}, opts ...DecoderConfigOption) error {
	err := decode(v.AllSettings(), defaultDecoderConfig(rawVal, opts...))

	if err != nil {
		return err
	}

	return nil
}

func (v *viperView) UnmarshalExact(rawVal interface{}) error {
	config := defaultDecoderConfig(rawVal)
	config.ErrorUnused = true

	err := decode(v.AllSettings(), config)

	if err != nil {
		return err
	}

	return nil
}
