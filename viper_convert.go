package viper

import "github.com/spf13/viper/internal/convert"

//MapTo quick map to struct if know what the value carries
//using `viper:"key"`` tag to specify keys
/*
	EG:
	type Service struct {
		Port int    `viper:"port"`
		IP   string `viper:"ip"`
	}

	SetDefault("service", map[string]interface{}{
		"ip":   "127.0.0.1",
		"port": 1234,
	})

	var service Service
	err := MapTo("service", &service)
	assert.NoError(t, err)
	assert.Equal(t, Get("service.port"), service.Port)
	assert.Equal(t, Get("service.ip"), service.IP)
*/
func MapTo(key string, target interface{}) error {
	return v.MapTo(key, target)
}

func (v *Viper) MapTo(key string, target interface{}) error {
	return convert.Convert(v.Get(key), target)
}
