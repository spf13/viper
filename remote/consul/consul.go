package consul

import (
	"bytes"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"github.com/spf13/viper"
	"io"
)

type consulConfigProvider struct{}

func (rc consulConfigProvider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	config := api.DefaultConfig()
	config.Address = rp.Endpoint()
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	kv, _, err := client.KV().Get(rp.Path(), nil)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(kv.Value), nil
}

func (rc consulConfigProvider) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	// TODO same as Get(), but behave like before(viper/remote), maybe record LastIndex in rp?
	resp, quit := newWatcher(rp.Path(), rp.Endpoint())
	r := <-resp
	close(quit)
	return bytes.NewReader(r.Value), r.Error
}

func (rc consulConfigProvider) WatchChannel(
	rp viper.RemoteProvider) (resp <-chan *viper.RemoteResponse, quit chan bool) {
	return newWatcher(rp.Path(), rp.Endpoint())
}

// To stop watch, just close(quit)
func newWatcher(key, addr string) (<-chan *viper.RemoteResponse, chan bool) {
	p, err := watch.Parse(map[string]interface{}{"type": "key", "key": key})
	if err != nil {
		return nil, nil
	}
	quit := make(chan bool)
	viperResponseCh := make(chan *viper.RemoteResponse)
	p.Handler = func(index uint64, data interface{}) {
		if data == nil {
			return
		}
		kv, ok := data.(*api.KVPair)
		if !ok {
			return
		}
		select {
		case viperResponseCh <- &viper.RemoteResponse{Value: kv.Value}:
		case <-quit:
		}
	}
	go p.Run(addr)
	// wait quit
	go func() {
		<-quit
		p.Stop()
	}()
	return viperResponseCh, quit
}

func init() {
	viper.RemoteConfig = &consulConfigProvider{}
}
