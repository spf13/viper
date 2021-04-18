module github.com/spf13/viper

go 1.12

require (
	github.com/bketelsen/crypt v0.0.3-0.20200106085610-5cbc8cc4026c
	github.com/fsnotify/fsnotify v1.4.7
	github.com/hashicorp/hcl v1.0.0
	github.com/magiconair/properties v1.8.1
	github.com/mitchellh/mapstructure v1.3.3
	github.com/pelletier/go-toml v1.2.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/afero v1.5.1
	github.com/spf13/cast v1.3.0
	github.com/spf13/jwalterweatherman v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/subosito/gotenv v1.2.0
	gopkg.in/ini.v1 v1.51.0
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/bketelsen/crypt => github.com/sagikazarmark/crypt v0.0.4-0.20210416192850-0e9f9535314d
