viper
=====

Go configuration with fangs

## What is Viper?

Viper is a complete configuration solution. Designed to work within an
application to handle file based configuration and seamlessly marry that with
command line flags which can also be used to control application behavior.

## Why Viper?

When building a modern application you don’t want to have to worry about
configuration file formats, you want to focus on building awesome software.
Viper is here to help with that.

Viper does the following for you:

1. Find, load and marshall a configuration file in YAML, TOML or JSON.
2. Provide a mechanism to setDefault values for your different configuration options
3. Provide a mechanism to setOverride values for options specified through command line flags.
4. Provide an alias system to easily rename parameters without breaking existing code.
5. Make it easy to tell the difference between when a user has provided a command line or config file which is the same as the default.

Viper believes that:

1. command line flags take precedence over options set in config files
2. config files take precedence over defaults

Config files often can be found in multiple locations. Viper allows you to set
multiple paths to search for the config file in.

Viper configuration keys are case insensitive.

## Usage

### Initialization

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath("/etc/appname/")   // path to look for the config file in
	viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
	viper.ReadInConfig() // Find and read the config file

### Setting Defaults

	viper.SetDefault("ContentDir", "content")
	viper.SetDefault("LayoutDir", "layouts")
	viper.SetDefault("Indexes", map[string]string{"tag": "tags", "category": "categories"})

### Setting Overrides

    viper.Set("Verbose", true)
    viper.Set("LogFile", LogFile)

### Registering and Using Aliases

    viper.RegisterAlias("loud", "Verbose")

    viper.Set("verbose", true) // same result as next line
    viper.Set("loud", true)   // same result as prior line

    viper.GetBool("loud") // true
    viper.GetBool("verbose") // true

### Getting Values

    viper.GetString("logfile") // case insensitive Setting & Getting
	if viper.GetBool("verbose") {
        fmt.Println("verbose enabled")
	}


## Q & A

Q: Why not INI files?

A: Ini files are pretty awful. There’s no standard format and they are hard to
validate. Viper is designed to work with YAML, TOML or JSON files. If someone
really wants to add this feature, I’d be happy to merge it. It’s easy to
specify which formats your application will permit.

Q: Why is it called "viper"?

A: Viper is designed to be a companion to
[Cobra](http://github.com/spf13/cobra). While both can operate completely
independently, together they make a powerful pair to handle much of your
application foundation needs.

Q: Why is it called "Cobra"?

A: Is there a better name for a commander?


