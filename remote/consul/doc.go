package consul

// Integrates the consul's remote features of Viper.

// viper/remote's problems:
// 1.consul watch do not work
// 2.outdated consul's lib

// use this ConsulRemoteConfigProvider, just replace
// _ "github.com/spf13/viper/remote" => _ "github.com/spf13/viper/remote/consul"

// usage example 1:
//func main() {
//	v := viper.New()
//	err := v.AddRemoteProvider("consul", "127.0.0.1:8500", "foo/bar")
//	if err != nil {
//		panic(err)
//	}
//	v.SetConfigType("json")
//	err = v.ReadRemoteConfig()
//	if err != nil {
//		panic(err)
//	}
//	for {
//		// blocking until consul kv updated
//		err = v.WatchRemoteConfig()
//		if err != nil {
//			panic(err)
//		}
//		// your config is updated now.
//		// ...
//	}
//}

// usage example 2:
//func main() {
//	v := viper.New()
//	err := v.AddRemoteProvider("consul", "127.0.0.1:8500", "foo/bar")
//	if err != nil {
//		panic(err)
//	}
//	v.SetConfigType("json")
//	err = v.ReadRemoteConfig()
//	if err != nil {
//		panic(err)
//	}
//	// your config will be update async
//	v.WatchRemoteConfigOnChannel()
//	// ....
//}