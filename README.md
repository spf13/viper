viper [![Build Status](https://travis-ci.org/spf13/viper.svg)](https://travis-ci.org/spf13/viper)
=====

[![Join the chat at https://gitter.im/spf13/viper](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/spf13/viper?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Go configuration with fangs

## What is Viper?

Viper is a complete configuration solution for go applications. It has
been designed to work within an application to handle all types of
configuration. It supports

* setting defaults
* reading from json, toml and yaml config files
* reading from environment variables
* reading from remote config systems (Etcd or Consul), watching changes
* reading from command line flags
* reading from buffer
* setting explicit values

It can be thought of as a registry for all of your applications
configuration needs.

## Why Viper?

When building a modern application, you don’t want to have to worry about
configuration file formats; you want to focus on building awesome software.
Viper is here to help with that.

Viper does the following for you:

1. Find, load and marshal a configuration file in JSON, TOML or YAML.
2. Provide a mechanism to set default values for your different
   configuration options.
3. Provide a mechanism to set override values for options specified
   through command line flags.
4. Provide an alias system to easily rename parameters without breaking
   existing code.
5. Make it easy to tell the difference between when a user has provided
   a command line or config file which is the same as the default.

Viper uses the following precedence order. Each item takes precedence
over the item below it:

 * explicit call to Set
 * flag
 * env
 * config
 * key/value store
 * default

Viper configuration keys are case insensitive.

## Putting Values into Viper

### Establishing Defaults

A good configuration system will support default values. A default value
is not required for a key, but can establish a default to be used in the
event that the key hasn’t be set via config file, environment variable,
remote configuration or flag.

Examples:

	viper.SetDefault("ContentDir", "content")
	viper.SetDefault("LayoutDir", "layouts")
	viper.SetDefault("Taxonomies", map[string]string{"tag": "tags", "category": "categories"})

### Reading Config Files

If you want to support a config file, Viper requires a minimal
configuration so it knows where to look for the config file. Viper
supports json, toml and yaml files. Viper can search multiple paths, but
currently a single viper only supports a single config file.

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath("/etc/appname/")   // path to look for the config file in
	viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
	err := viper.ReadInConfig() // Find and read the config file
    if err != nil { // Handle errors reading the config file
        panic(fmt.Errorf("Fatal error config file: %s \n", err))
    }

### Reading Config from io.Reader

Viper predefined many configuration sources, such as files, environment variables, flags and 
remote K/V store. But you are not bound to them. You can also implement your own way to
require configuration and feed it to viper.

````go
viper.SetConfigType("yaml") // or viper.SetConfigType("YAML")

// any approach to require this configuration into your program. 
var yamlExample = []byte(`
Hacker: true
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

viper.ReadConfig(bytes.NewBuffer(yamlExample))

viper.Get("name") // this would be "steve"
````

### Setting Overrides

These could be from a command line flag, or from your own application logic.

    viper.Set("Verbose", true)
    viper.Set("LogFile", LogFile)

### Registering and Using Aliases

Aliases permit a single value to be referenced by multiple keys

    viper.RegisterAlias("loud", "Verbose")

    viper.Set("verbose", true) // same result as next line
    viper.Set("loud", true)   // same result as prior line

    viper.GetBool("loud") // true
    viper.GetBool("verbose") // true

### Working with Environment Variables

Viper has full support for environment variables. This enables 12 factor
applications out of the box. There are four methods that exist to aid
with working with ENV:

 * AutomaticEnv()
 * BindEnv(string...) : error
 * SetEnvPrefix(string)
 * SetEnvReplacer(string...) *strings.Replacer

_When working with ENV variables, it’s important to recognize that Viper
treats ENV variables as case sensitive._

Viper provides a mechanism to try to ensure that ENV variables are
unique. By using SetEnvPrefix, you can tell Viper to use add a prefix
while reading from the environment variables. Both BindEnv and
AutomaticEnv will use this prefix.

BindEnv takes one or two parameters. The first parameter is the key
name, the second is the name of the environment variable. The name of
the environment variable is case sensitive. If the ENV variable name is
not provided, then Viper will automatically assume that the key name
matches the ENV variable name but the ENV variable is IN ALL CAPS. When
you explicitly provide the ENV variable name, it **does not**
automatically add the prefix.

One important thing to recognize when working with ENV variables is that
the value will be read each time it is accessed. It does not fix the
value when the BindEnv is called.

AutomaticEnv is a powerful helper especially when combined with
SetEnvPrefix. When called, Viper will check for an environment variable
any time a viper.Get request is made. It will apply the following rules.
It will check for a environment variable with a name matching the key
uppercased and prefixed with the EnvPrefix if set.

SetEnvReplacer allows you to use a `strings.Replacer` object to rewrite Env keys
to an extent. This is useful if you want to use `-` or something in your Get()
calls, but want your environmental variables to use `_` delimiters. An example
of using it can be found in `viper_test.go`.

#### Env example

	SetEnvPrefix("spf") // will be uppercased automatically
	BindEnv("id")

	os.Setenv("SPF_ID", "13") // typically done outside of the app

	id := Get("id") // 13


### Working with Flags

Viper has the ability to bind to flags. Specifically, Viper supports
Pflags as used in the [Cobra](https://github.com/spf13/cobra) library.

Like BindEnv, the value is not set when the binding method is called, but
when it is accessed. This means you can bind as early as you want, even
in an init() function.

The BindPFlag() method provides this functionality.

Example:

    serverCmd.Flags().Int("port", 1138, "Port to run Application server on")
    viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))


### Remote Key/Value Store Support

To enable remote support in Viper, do a blank import of the `viper/remote` package:

`import _ github.com/spf13/viper/remote`

Viper will read a config string (as JSON, TOML, or YAML) retrieved from a
path in a Key/Value store such as Etcd or Consul.  These values take precedence
over default values, but are overriden by configuration values retrieved from disk,
flags, or environment variables.

Viper uses [crypt](https://github.com/xordataexchange/crypt) to retrieve configuration
from the K/V store, which means that you can store your configuration values
encrypted and have them automatically decrypted if you have the correct
gpg keyring.  Encryption is optional.

You can use remote configuration in conjunction with local configuration, or
independently of it.

`crypt` has a command-line helper that you can use to put configurations
in your K/V store. `crypt` defaults to etcd on http://127.0.0.1:4001.

	go get github.com/xordataexchange/crypt/bin/crypt
	crypt set -plaintext /config/hugo.json /Users/hugo/settings/config.json

Confirm that your value was set:

	crypt get -plaintext /config/hugo.json

See the `crypt` documentation for examples of how to set encrypted values, or how
to use Consul.

### Remote Key/Value Store Example - Unencrypted

	viper.AddRemoteProvider("etcd", "http://127.0.0.1:4001","/config/hugo.json")
	viper.SetConfigType("json") // because there is no file extension in a stream of bytes
	err := viper.ReadRemoteConfig()

### Remote Key/Value Store Example - Encrypted

	viper.AddSecureRemoteProvider("etcd","http://127.0.0.1:4001","/config/hugo.json","/etc/secrets/mykeyring.gpg")
	viper.SetConfigType("json") // because there is no file extension in a stream of bytes
	err := viper.ReadRemoteConfig()

### Watching Changes in Etcd - Unencrypted

    // alternatively, you can create a new viper instance.
    var runtime_viper = viper.New()

    runtime_viper.AddRemoteProvider("etcd", "http://127.0.0.1:4001", "/config/hugo.yml")
    runtime_viper.SetConfigType("yaml") // because there is no file extension in a stream of bytes

    // read from remote config the first time.
    err := runtime_viper.ReadRemoteConfig()

    // marshal config
    runtime_viper.Marshal(&runtime_conf)

    // open a goroutine to wath remote changes forever
    go func(){
        for {
            time.Sleep(time.Second * 5) // delay after each request

            // currenlty, only tested with etcd support
            err := runtime_viper.WatchRemoteConfig()
            if err != nil {
                log.Errorf("unable to read remote config: %v", err)
                continue
            }

            // marshal new config into our runtime config struct. you can also use channel 
            // to implement a signal to notify the system of the changes
            runtime_viper.Marshal(&runtime_conf)
        }
    }()


## Getting Values From Viper

In Viper, there are a few ways to get a value depending on what type of value you want to retrieved.
The following functions and methods exist:

 * Get(key string) : interface{}
 * GetBool(key string) : bool
 * GetFloat64(key string) : float64
 * GetInt(key string) : int
 * GetString(key string) : string
 * GetStringMap(key string) : map[string]interface{}
 * GetStringMapString(key string) : map[string]string
 * GetStringSlice(key string) : []string
 * GetTime(key string) : time.Time
 * GetDuration(key string) : time.Duration
 * IsSet(key string) : bool

One important thing to recognize is that each Get function will return
its zero value if it’s not found. To check if a given key exists, the IsSet()
method has been provided.

Example:

    viper.GetString("logfile") // case-insensitive Setting & Getting
    if viper.GetBool("verbose") {
        fmt.Println("verbose enabled")
    }

### Accessing nested keys

The accessor methods also accept formatted paths to deeply nested keys. 
For example, if the following JSON file is loaded:

```
{
    "host": {
        "address": "localhost",
        "port": 5799
    },
    "datastore": {
        "metric": {
            "host": "127.0.0.1",
            "port": 3099
        },
        "warehouse": {
            "host": "198.0.0.1",
            "port": 2112
        }
    }
}

```

Viper can access a nested field by passing a `.` delimited path of keys:
```
GetString("datastore.metric.host") // (returns "127.0.0.1")
```

This obeys the precendense rules established above; the search for the root key
(in this examole, `datastore`) will cascade through the remaining configuration registries
until found. The search for the subkeys (`metric` and `host`), however, will not.

For example, if the `metric` key was not defined in the configuration loaded
from file, but was defined in the defaults, Viper would return the zero value.

On the other hand, if the primary key was not defined, Viper would go through the
remaining registries looking for it.

Lastly, if there exists a key that matches the delimited key path, its value will
be returned instead. E.g. 

```
{
    "datastore.metric.host": "0.0.0.0",
    "host": {
        "address": "localhost",
        "port": 5799
    },
    "datastore": {
        "metric": {
            "host": "127.0.0.1",
            "port": 3099
        },
        "warehouse": {
            "host": "198.0.0.1",
            "port": 2112
        }
    }
}

GetString("datastore.metric.host") //returns "0.0.0.0"
```

### Marshaling

You also have the option of Marshaling all or a specific value to a struct, map, etc.

There are two methods to do this:

 * Marshal(rawVal interface{}) : error
 * MarshalKey(key string, rawVal interface{}) : error

Example:

	type config struct {
		Port int
		Name string
	}

	var C config

	err := Marshal(&C)
	if err != nil {
		t.Fatalf("unable to decode into struct, %v", err)
	}


## Viper or Vipers?

Viper comes ready to use out of the box. There is no configuration or
initialization needed to begin using Viper. Since most applications will
want to use a single central repository for their configuration, the
viper package provides this. It is similar to a singleton.

In all of the examples above, they demonstrate using viper in its
singleton style approach.

### Working with multiple vipers

You can also create many different vipers for use in your application.
Each will have it’s own unique set of configurations and values. Each
can read from a different config file, key value store, etc. All of the
functions that viper package supports are mirrored as methods on a viper.

Example:

    x := viper.New()
    y := viper.New()

	x.SetDefault("ContentDir", "content")
	y.SetDefault("ContentDir", "foobar")

    ...

When working with multiple vipers, it is up to the user to keep track of
the different vipers.

## Q & A

Q: Why not INI files?

A: Ini files are pretty awful. There’s no standard format, and they are hard to
validate. Viper is designed to work with JSON, TOML or YAML files. If someone
really wants to add this feature, I’d be happy to merge it. It’s easy to
specify which formats your application will permit.

Q: Why is it called “Viper”?

A: Viper is designed to be a [companion](http://en.wikipedia.org/wiki/Viper_(G.I._Joe)) to
[Cobra](https://github.com/spf13/cobra). While both can operate completely
independently, together they make a powerful pair to handle much of your
application foundation needs.

Q: Why is it called “Cobra”?

A: Is there a better name for a [commander](http://en.wikipedia.org/wiki/Cobra_Commander)?
