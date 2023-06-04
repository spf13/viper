package viper

type RegisteredConfig struct {
	Key            string
	CanBeNil       bool
	OnUpdate       func(e *Event)
	OnUpdateFailed func(e *Event)
	Schema         interface{}
	Validator      func(interface{}) bool
}

func (v *Viper) Register(r []RegisteredConfig) {
	if v.registered == nil {
		v.registered = make(map[string]RegisteredConfig)
	}
	for _, config := range r {
		v.registered[config.Key] = config
	}
}
