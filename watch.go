//go:build darwin || dragonfly || freebsd || openbsd || linux || netbsd || solaris || windows
// +build darwin dragonfly freebsd openbsd linux netbsd solaris windows

package viper

import "github.com/fsnotify/fsnotify"

type watcher = fsnotify.Watcher

func newWatcher() (*watcher, error) {
	return fsnotify.NewWatcher()
}

type Event struct {
	new interface{}
	old interface{}
}

func (s *Event) New() interface{} {
	return s.new
}

func (s *Event) Old() interface{} {
	return s.old
}
