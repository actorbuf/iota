package etcd

import (
	"fmt"
	"testing"
)

var (
	testServer = []string{"10.0.0.94:32349"}
	testKey    = []string{"/test/hello", "/test/world"}
)

func TestEtcdWatcher_AddWatch(t *testing.T) {
	watcher, err := NewEtcdWatcher(testServer)
	if err != nil {
		panic(err)
	}
	for _, key := range testKey {
		watcher.AddWatch(key, false, emptyListen{})
	}
	select {}
}

type emptyListen struct{}

func (e emptyListen) Set(key []byte, value []byte, version int64) {
	fmt.Println("set_key:", string(key), "set_val:", string(value), "version:", version)
}

func (e emptyListen) Create(key []byte, value []byte, version int64) {
	fmt.Println("create_key:", string(key), "create_val:", string(value), "version:", version)
}

func (e emptyListen) Modify(key []byte, value []byte, version int64) {
	fmt.Println("modify_key:", string(key), "modify_val:", string(value), "version:", version)
}

func (e emptyListen) Delete(key []byte, version int64) {
	fmt.Println("delete:", string(key), "version:", version)
}

func (e emptyListen) Exit(err string) {
	fmt.Println("watch exit", err)
}
