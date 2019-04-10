package consul

import (
	"bytes"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"github.com/spf13/viper"
	"io"
	"sync"
)

type consulConfigProvider struct {
	mu     sync.Mutex
	idxMap map[string]uint64
}

func (rc *consulConfigProvider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	config := api.DefaultConfig()
	config.Address = rp.Endpoint()
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	kv, meta, err := client.KV().Get(rp.Path(), nil)
	if err != nil {
		return nil, err
	}
	rc.updateIndex(rp, meta.LastIndex)
	return bytes.NewReader(kv.Value), nil
}

func (rc *consulConfigProvider) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	return rc.watch(rp)
}

func (rc *consulConfigProvider) watch(rp viper.RemoteProvider) (r io.Reader, err error) {
	p, err := watch.Parse(map[string]interface{}{"type": "key", "key": rp.Path()})
	if err != nil {
		return nil, err
	}
	// handler
	p.Handler = func(index uint64, data interface{}) {
		if data == nil {
			return
		}
		kv, ok := data.(*api.KVPair)
		if !ok {
			return
		}
		if !rc.updateIndex(rp, index) {
			return
		}
		r = bytes.NewReader(kv.Value)
		p.Stop()
	}
	// start watch
	p.Run(rp.Endpoint())
	return
}

func (rc *consulConfigProvider) WatchChannel(
	rp viper.RemoteProvider) (resp <-chan *viper.RemoteResponse, quit chan bool) {
	return rc.watchChannel(rp)
}

func (rc *consulConfigProvider) watchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	p, err := watch.Parse(map[string]interface{}{"type": "key", "key": rp.Path()})
	if err != nil {
		// this should not happen
		return nil, nil
	}
	quit := make(chan bool)
	viperResponseCh := make(chan *viper.RemoteResponse)
	// handler
	p.Handler = func(index uint64, data interface{}) {
		if data == nil {
			return
		}
		kv, ok := data.(*api.KVPair)
		if !ok {
			return
		}
		if !rc.updateIndex(rp, index) {
			return
		}
		select {
		case viperResponseCh <- &viper.RemoteResponse{Value: kv.Value}:
		case <-quit:
		}
	}
	// start watcher
	go p.Run(rp.Endpoint())
	// wait quit
	go func() {
		<-quit
		p.Stop()
	}()
	return viperResponseCh, quit
}

func makeIndexKey(rp viper.RemoteProvider) string {
	return rp.Endpoint() + "_" + rp.Path()
}

func (rc *consulConfigProvider) updateIndex(
	rp viper.RemoteProvider, lastIndex uint64) (updated bool) {
	rc.mu.Lock()
	oldLastIndex := rc.idxMap[makeIndexKey(rp)]
	if oldLastIndex < lastIndex {
		rc.idxMap[makeIndexKey(rp)] = lastIndex
		updated = true
	}
	rc.mu.Unlock()
	return
}

func init() {
	viper.RemoteConfig = &consulConfigProvider{idxMap: make(map[string]uint64)}
}
