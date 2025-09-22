> ## Viper v2 Feedback
> Viper is heading towards v2 and we would love to hear what _**you**_ would like to see in it. Share your thoughts here: https://forms.gle/R6faU74qPRPAzchZ9
>
> **Thank you!**

![viper logo](https://github.com/user-attachments/assets/acae9193-2974-41f3-808d-2d433f5ada5e)


[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go#configuration)
[![run on repl.it](https://repl.it/badge/github/sagikazarmark/Viper-example)](https://repl.it/@sagikazarmark/Viper-example#main.go)

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/spf13/viper/ci.yaml?branch=master&style=flat-square)](https://github.com/spf13/viper/actions?query=workflow%3ACI)
[![Join the chat at https://gitter.im/spf13/viper](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/spf13/viper?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/spf13/viper?style=flat-square)](https://goreportcard.com/report/github.com/spf13/viper)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.23-61CFDD.svg?style=flat-square)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/spf13/viper)](https://pkg.go.dev/mod/github.com/spf13/viper)

**Go configuration with fangs!**

Many Go projects are built using Viper including:

* [Hugo](http://gohugo.io)
* [EMC RexRay](http://rexray.readthedocs.org/en/stable/)
* [Imgur’s Incus](https://github.com/Imgur/incus)
* [Nanobox](https://github.com/nanobox-io/nanobox)/[Nanopack](https://github.com/nanopack)
* [Docker Notary](https://github.com/docker/Notary)
* [BloomApi](https://www.bloomapi.com/)
* [doctl](https://github.com/digitalocean/doctl)
* [Clairctl](https://github.com/jgsqware/clairctl)
* [Mercure](https://mercure.rocks)
* [Meshery](https://github.com/meshery/meshery)
* [Bearer](https://github.com/bearer/bearer)
* [Coder](https://github.com/coder/coder)
* [Vitess](https://vitess.io/)


## Install

```shell
go get github.com/spf13/viper
```

> **NOTE** Viper uses [Go Modules](https://go.dev/wiki/Modules) to manage dependencies.


## Why use Viper?

Viper is a complete configuration solution for Go applications including
[12-Factor apps](https://12factor.net/#the_twelve_factors). It is designed to
work within any application, and can handle all types of configuration needs
and formats. It supports:

* setting defaults
* setting explicit values
* reading config files
* dynamic discovery of config files across multiple locations
* reading from environment variables
* reading from remote systems (e.g. Etcd or Consul)
* reading from command line flags
* reading from buffers
* live watching and updating configuration
* aliasing configuration keys for easy refactoring

Viper can be thought of as a registry for all of your applications'
configuration needs.


## Putting Values in Viper

Viper can read from multiple configuration sources and merges them together
into one set of configuration keys and values.

Viper uses the following precedence for merging:

 * explicit call to `Set`
 * flags
 * environment variables
 * config files
 * external key/value stores
 * defaults

> **NOTE** Viper configuration keys are case insensitive.

### Reading Config Files

Viper requires minimal configuration to load config files. Viper currently supports:

* JSON
* TOML
* YAML
* HCL
* INI
* envfile
* Java Propeties

A single Viper instance only supports a single configuration file, but multiple
paths may be searched for one.

Here is an example of how to use Viper to search for and read a configuration
file. At least one path should be provided where a configuration file is
expected.

```go
// Name of the config file without an extension (Viper will intuit the type
// from an extension on the actual file)
viper.SetConfigName("config")

// Add search paths to find the file
viper.AddConfigPath("/etc/appname/")
viper.AddConfigPath("$HOME/.appname")
viper.AddConfigPath(".")

// Find and read the config file
err := viper.ReadInConfig()

// Handle errors
if err != nil {
	panic(fmt.Errorf("fatal error config file: %w", err))
}
```

You can handle the specific case where no config file is found.

```go
if err := viper.ReadInConfig(); err != nil {
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found; ignore error if desired
	} else {
		// Config file was found but another error was produced
	}
}

// Config file found and successfully parsed
```

> **NOTE (since 1.6)** You can also have a file without an extension and
> specify the format programmatically, which is useful for files that naturally
> have no extension (e.g., `.bashrc`).

### Writing Config Files

At times you may want to store all configuration modifications made during run
time.

```go
// Writes current config to the path set by `AddConfigPath` and `SetConfigName`
viper.WriteConfig()
viper.SafeWriteConfig() // Like the above, but will error if the config file exists

// Writes current config to a specific place
viper.WriteConfigAs("/path/to/my/.config")

// Will error since it has already been written
viper.SafeWriteConfigAs("/path/to/my/.config")

viper.SafeWriteConfigAs("/path/to/my/.other_config")
```

As a rule of the thumb, methods prefixed with `Safe` won't overwrite any
existing file, while other methods will.

### Watching and Re-reading Config Files

Gone are the days of needing to restart a server to have a config take
effect--Viper powered applications can read an update to a config file while
running and not miss a beat.

It's also possible to provide a function for Viper to run each time a change
occurs.

```go
// All config paths must be defined prior to calling `WatchConfig()`
viper.AddConfigPath("$HOME/.appname")

viper.OnConfigChange(func(e fsnotify.Event) {
	fmt.Println("Config file changed:", e.Name)
})

viper.WatchConfig()
```

### Reading Config from `io.Reader`

Viper predefines many configuration sources but you can also implement your own
required configuration source.

```go
viper.SetConfigType("yaml")

var yamlExample = []byte(`
hacker: true
hobbies:
- skateboarding
- snowboarding
- go
name: steve
`)

viper.ReadConfig(bytes.NewBuffer(yamlExample))

viper.Get("name") // "steve"
```

### Setting Defaults

A good configuration system will support default values, which are used if a
key hasn't been set in some other way.

Examples:

```go
viper.SetDefault("ContentDir", "content")
viper.SetDefault("LayoutDir", "layouts")
viper.SetDefault("Taxonomies", map[string]string{"tag": "tags", "category": "categories"})
```

### Setting Overrides

Viper allows explict setting of configuration, such as from your own
application logic.

```go
viper.Set("verbose", true)
viper.Set("host.port", 5899) // Set an embedded key
```

### Registering and Using Aliases

Aliases permit a single value to be referenced by multiple keys

```go
viper.RegisterAlias("loud", "Verbose")

viper.Set("verbose", true) // Same result as next line
viper.Set("loud", true)    // Same result as prior line

viper.GetBool("loud")    // true
viper.GetBool("verbose") // true
```

### Working with Environment Variables

Viper has full support for environment variables.

> **NOTE** Unlike other configuration sources, environment variables are case
> sensitive.

```go
// Tells Viper to use this prefix when reading environment variables
viper.SetEnvPrefix("spf")

// Viper will look for "SPF_ID", automatically uppercasing the prefix and key
viper.BindEnv("id")

// Alternatively, we can search for any environment variable prefixed and load
// them in
viper.AutomaticEnv()

os.Setenv("SPF_ID", "13")

id := viper.Get("id") // 13
```

* By default, empty environment variables are considered unset and will fall back to
  the next configuration source, unless `AllowEmptyEnv` is used.
* Viper does not "cache" environment variables--the value will be read each
  time it is accessed.
* `SetEnvKeyReplacer` and `EnvKeyReplacer` allow you to use to rewrite Env
  keys, which is useful to combine SCREAMING_SNAKE_CASE environment variables
  to merge with kebab-cased configuration values from other sources.

### Working with Flags

Viper has the ability to bind to flags. Specifically, Viper supports
[pflag](https://github.com/spf13/pflag/) as used in the
[Cobra](https://github.com/spf13/cobra) library.

Like environment variables, the value is not set when the binding method is
called, but when it is accessed.

For individual flags, the `BindPFlag` method provides this functionality.

```go
serverCmd.Flags().Int("port", 1138, "Port to run Application server on")

viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))
```

You can also bind an existing set of pflags.

```go
pflag.Int("flagname", 1234, "help message for flagname")
pflag.Parse()

viper.BindPFlags(pflag.CommandLine)

i := viper.GetInt("flagname") // Retrieve values from viper instead of pflag
```

The standard library [flag](https://golang.org/pkg/flag/) package is not
directly supported, but may be parsed through pflag.

```go
package main

import (
	"flag"

	"github.com/spf13/pflag"
)

func main() {
	// Using standard library "flag" package
	flag.Int("flagname", 1234, "help message for flagname")

    // Pass standard library flags to pflag
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

    // Viper takes over
	viper.BindPFlags(pflag.CommandLine)
}
```

Use of pflag may be avoided entirely by implementing the `FlagValue` and
`FlagValueSet` interfaces.

```go
// Implementing FlagValue

type myFlag struct {}
func (f myFlag) HasChanged() bool { return false }
func (f myFlag) Name() string { return "my-flag-name" }
func (f myFlag) ValueString() string { return "my-flag-value" }
func (f myFlag) ValueType() string { return "string" }

viper.BindFlagValue("my-flag-name", myFlag{})

// Implementing FlagValueSet

type myFlagSet struct {
	flags []myFlag
}
func (f myFlagSet) VisitAll(fn func(FlagValue)) {
	for _, flag := range flags {
		fn(flag)
	}
}

fSet := myFlagSet{
	flags: []myFlag{myFlag{}, myFlag{}},
}
viper.BindFlagValues("my-flags", fSet)
```

### Remote Key/Value Store Support

To enable remote support in Viper, do a blank import of the `viper/remote`
package.

```go
import _ "github.com/spf13/viper/remote"
```

Viper supports the following remote key/value stores. Examples for each are
provided below.

* Etcd and Etcd3
* Consul
* Firestore
* NATS

Viper will read a config string (as JSON, TOML, YAML, HCL or envfile) retrieved
from a path in a key/value store.

Viper supports multiple hosts separated by `;`. For example:
`http://127.0.0.1:4001;http://127.0.0.1:4002`.

#### Encryption

Viper uses [crypt](https://github.com/sagikazarmark/crypt) to retrieve
configuration from the key/value store, which means that you can store your
configuration values encrypted and have them automatically decrypted if you
have the correct GPG keyring. Encryption is optional.

Crypt has a command-line helper that you can use to put configurations in your
key/value store.

```bash
$ go get github.com/sagikazarmark/crypt/bin/crypt
$ crypt set -plaintext /config/hugo.json /Users/hugo/settings/config.json
$ crypt get -plaintext /config/hugo.json
```

See the Crypt documentation for examples of how to set encrypted values, or
how to use Consul.

### Remote Key/Value Store Examples (unencrypted)

#### etcd

```go
viper.AddRemoteProvider("etcd", "http://127.0.0.1:4001","/config/hugo.json")
viper.SetConfigType("json") // because there is no file extension in a stream of bytes, supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop", "env", "dotenv"
err := viper.ReadRemoteConfig()
```

#### etcd3

```go
viper.AddRemoteProvider("etcd3", "http://127.0.0.1:4001","/config/hugo.json")
viper.SetConfigType("json") // because there is no file extension in a stream of bytes, supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop", "env", "dotenv"
err := viper.ReadRemoteConfig()
```

#### Consul

Given a Consul key `MY_CONSUL_KEY` with the value:

```json
{
    "port": 8080,
    "hostname": "myhostname.com"
}
```

```go
viper.AddRemoteProvider("consul", "localhost:8500", "MY_CONSUL_KEY")
viper.SetConfigType("json") // Need to explicitly set this to json
err := viper.ReadRemoteConfig()

fmt.Println(viper.Get("port")) // 8080
```

#### Firestore

```go
viper.AddRemoteProvider("firestore", "google-cloud-project-id", "collection/document")
viper.SetConfigType("json") // Config's format: "json", "toml", "yaml", "yml"
err := viper.ReadRemoteConfig()
```

Of course, you're allowed to use `SecureRemoteProvider` also.

#### NATS

```go
viper.AddRemoteProvider("nats", "nats://127.0.0.1:4222", "myapp.config")
viper.SetConfigType("json")
err := viper.ReadRemoteConfig()
```

### Remote Key/Value Store Examples (encrypted)

```go
viper.AddSecureRemoteProvider("etcd","http://127.0.0.1:4001","/config/hugo.json","/etc/secrets/mykeyring.gpg")
viper.SetConfigType("json") // because there is no file extension in a stream of bytes,  supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop", "env", "dotenv"
err := viper.ReadRemoteConfig()
```

### Watching Key/Value Store Changes

```go
// Alternatively, you can create a new viper instance
var runtime_viper = viper.New()

runtime_viper.AddRemoteProvider("etcd", "http://127.0.0.1:4001", "/config/hugo.yml")
runtime_viper.SetConfigType("yaml") // because there is no file extension in a stream of bytes, supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop", "env", "dotenv"

// Read from remote config the first time
err := runtime_viper.ReadRemoteConfig()

// Unmarshal config
runtime_viper.Unmarshal(&runtime_conf)

// Open a goroutine to watch remote changes forever
go func(){
	for {
		time.Sleep(time.Second * 5) // delay after each request

		// Currently, only tested with Etcd support
		err := runtime_viper.WatchRemoteConfig()
		if err != nil {
			log.Errorf("unable to read remote config: %v", err)
			continue
		}

		// Unmarshal new config into our runtime config struct
		runtime_viper.Unmarshal(&runtime_conf)
	}
}()
```

## Getting Values From Viper

THe simplest way to retrieve configuration values from Viper is to use `Get*`
functions. `Get` will return an any type, but specific types may be retrieved
with `Get<Type>` functions.

Note that each `Get*` function will return a zero value if it’s key is not
found. To check if a key exists, use the `IsSet` method.

Nested keys use `.` as a delimiter and numbers for array indexes. Given the
following configuration:

```jsonc
{
    "datastore": {
        "metric": {
            "host": "127.0.0.1",
            "ports": [
                5799,
                6029
            ]
        }
    }
}

```

```go
GetString("datastore.metric.host") // "127.0.0.1"
GetInt("host.ports.1") // 6029
```

> **NOTE** Viper _does not_ deep merge configuration values. Complex values
> that are overridden will be entirely replaced.

If there exists a key that matches the delimited key path, its value will be
returned instead.

```jsonc
{
    "datastore.metric.host": "0.0.0.0",
    "datastore": {
        "metric": {
            "host": "127.0.0.1"
        }
    }
}
```

```go
GetString("datastore.metric.host") // "0.0.0.0"
```

### Configuration Subsets

It's often useful to extract a subset of configuration (e.g., when developing a
reusable module which should accept specific sections of configuration).

```yaml
cache:
  cache1:
    item-size: 64
    max-items: 100
  cache2:
    item-size: 80
    max-items: 200
```

```go
func NewCache(v *Viper) *Cache {
	return &Cache{
		ItemSize: v.GetInt("item-size"),
		MaxItems: v.GetInt("max-items"),
	}
}

cache1Config := viper.Sub("cache.cache1")

if cache1Config == nil {
    // Sub returns nil if the key cannot be found
	panic("cache configuration not found")
}

cache1 := NewCache(cache1Config)
```

### Unmarshaling

You also have the option of Unmarshaling all or a specific value to a struct,
map, and etc., using `Unmarshal*` methods.

```go
type config struct {
	Port int
	Name string
	PathMap string `mapstructure:"path_map"`
}

var C config

err := viper.Unmarshal(&C)
if err != nil {
	t.Fatalf("unable to decode into struct, %v", err)
}
```

If you want to unmarshal configuration where the keys themselves contain `.`
(the default key delimiter), you can change the delimiter.

```go
v := viper.NewWithOptions(viper.KeyDelimiter("::"))

v.SetDefault("chart::values", map[string]any{
	"ingress": map[string]any{
		"annotations": map[string]any{
			"traefik.frontend.rule.type":                 "PathPrefix",
			"traefik.ingress.kubernetes.io/ssl-redirect": "true",
		},
	},
})

type config struct {
	Chart struct{
		Values map[string]any
	}
}

var C config

v.Unmarshal(&C)
```

Viper also supports unmarshaling into embedded structs.

```go
/*
Example config:

module:
    enabled: true
    token: 89h3f98hbwf987h3f98wenf89ehf
*/
type config struct {
	Module struct {
		Enabled bool

		moduleConfig `mapstructure:",squash"`
	}
}

type moduleConfig struct {
	Token string
}

var C config

err := viper.Unmarshal(&C)
if err != nil {
	t.Fatalf("unable to decode into struct, %v", err)
}
```

Viper uses
[github.com/go-viper/mapstructure](https://github.com/go-viper/mapstructure)
under the hood for unmarshaling values which uses `mapstructure` tags, by
default.

### Marshalling to String

You may need to marshal all the settings held in viper into a string rather
than write them to a file. You can use your favorite format's marshaller with
the config returned by `AllSettings`.

```go
import (
	yaml "go.yaml.in/yaml/v3"
)

func yamlStringSettings() string {
	c := viper.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	return string(bs)
}
```

### Decoding Custom Formats

A frequently requested feature for Viper is adding more value formats and
decoders. For example, parsing character (dot, comma, semicolon, etc) separated
strings into slices. This is already available in Viper using mapstructure
decode hooks.

Read more in [this blog
post](https://sagikazarmark.hu/blog/decoding-custom-formats-with-viper/).


## FAQ

### Why is it called “Viper”?

Viper is designed to be a
[companion](http://en.wikipedia.org/wiki/Viper_(G.I._Joe)) to
[Cobra](https://github.com/spf13/cobra). While both can operate completely
independently, together they make a powerful pair to handle much of your
application foundation needs.

### I found a bug or want a feature, should I file an issue or a PR?

Yes, but there are two things to be aware of.

1.  The Viper project is currently prioritizing backwards compatibility and
    stability over features.
2.  Features may be deferred until Viper 2 forms.

### Can multiple Viper instances be used?

**tl;dr:** Yes.

Each will have its own unique configuration and can read from a different
configuration source. All of the functions that the Viper package supports are
mirrored as methods on a Viper instance.

```go
x := viper.New()
y := viper.New()

x.SetDefault("ContentDir", "content")
y.SetDefault("ContentDir", "foobar")
```

### Should Viper be a global singleton or passed around?

The best practice is to initialize a Viper instance and pass that around when
necessary.

Viper comes with a global instance (singleton) out of the box. Although it
makes setting up configuration easy, using it is generally discouraged as it
makes testing harder and can lead to unexpected behavior.

The global instance may be deprecated in the future. See
[#1855](https://github.com/spf13/viper/issues/1855) for more details.

### Does Viper support case sensitive keys?

**tl;dr:** No.

Viper merges configuration from various sources, many of which are either case
insensitive or use different casing than other sources (e.g., env vars). In
order to provide the best experience when using multiple sources, all keys are
made case insensitive.

There has been several attempts to implement case sensitivity, but
unfortunately it's not trivial. We might take a stab at implementing it in
[Viper v2](https://github.com/spf13/viper/issues/772), but despite the initial
noise, it does not seem to be requested that much.

You can vote for case sensitivity by filling out this feedback form:
https://forms.gle/R6faU74qPRPAzchZ9.

### Is it safe to concurrently read and write to a Viper instance?

No, you will need to synchronize access to Viper yourself (for example by using
the `sync` package). Concurrent reads and writes can cause a panic.


## Troubleshooting

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md).


## Development

**For an optimal developer experience, it is recommended to install
[Nix](https://nixos.org/download.html) and
[direnv](https://direnv.net/docs/installation.html).**

_Alternatively, install [Go](https://go.dev/dl/) on your computer then run
`make deps` to install the rest of the dependencies._

Run the test suite:

```shell
make test
```

Run linters:

```shell
make lint # pass -j option to run them in parallel
```

Some linter violations can automatically be fixed:

```shell
make fmt
```


## License

The project is licensed under the [MIT License](LICENSE).
