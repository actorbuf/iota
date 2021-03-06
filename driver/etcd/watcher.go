// Package etcd ETCD键值监听器
package etcd

import (
	"context"
	"errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

var (
	timeOut = time.Duration(3) * time.Second // 超时
	watcher *Watcher
)

// Listener 对外通知接口
// key是键（以字节为单位）。不允许使用空键。
// value是key保存的值，以字节为单位。
// version是key的版本version。删除会将version重置为零，并且对key的任何修改都会增加其version。
// 当Exit方法被调用时，说明watch已经退出。err为退出原因（如果有）。
type Listener interface {
	Set(key []byte, value []byte, version int64)
	Create(key []byte, value []byte, version int64)
	Modify(key []byte, value []byte, version int64)
	Delete(key []byte, version int64)
	Exit(err string)
}

type WatcherStarter struct {
	Servers []string
}

func GetEtcdWatcher() *Watcher {
	return watcher
}

// Watcher 是维护的一个ETCD key监视器
type Watcher struct {
	// etcd client
	cli *clientv3.Client
	// 监听事件对外通知具体实现
	listener Listener
	// 保护内部字段
	mu sync.Mutex
	// 关闭通知
	closeHandler map[string]func()
}

// NewEtcdWatcher 构造一个新的EtcdWatcher.
func NewEtcdWatcher(servers []string) (*Watcher, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:          servers,
		DialTimeout:        timeOut,
		MaxCallSendMsgSize: 11 * (1 << 20),
	})
	if err != nil {
		return nil, err
	}

	watcher = &Watcher{
		cli:          cli,
		closeHandler: make(map[string]func()),
	}

	return watcher, nil
}

// AddWatch 添加监视.
// key是需要监听的键.
// prefix允许对具有匹配前缀的键进行操作。例如，“ Get（foo，WithPrefix（））” 可以返回“ foo1”，“ foo2”，依此类推.
// listener当监听到对应的事件时，将动作转发到Listener对应的实现. 使用前，请先实现Listener接口.
func (mgr *Watcher) AddWatch(key string, prefix bool, listener Listener) bool {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	if _, ok := mgr.closeHandler[key]; ok {
		return false
	}
	ctx, cancel := context.WithCancel(context.Background())
	mgr.closeHandler[key] = cancel

	go func() {
		_ = mgr.watch(ctx, key, prefix, listener)
	}()

	return true
}

// RemoveWatch 删除监视
func (mgr *Watcher) RemoveWatch(key string) bool {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	cancel, ok := mgr.closeHandler[key]
	if !ok {
		return false
	}
	cancel()
	delete(mgr.closeHandler, key)

	return true
}

// ClearWatch 清除所有监视
func (mgr *Watcher) ClearWatch() {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	for k := range mgr.closeHandler {
		mgr.closeHandler[k]()
	}
	mgr.closeHandler = make(map[string]func())
}

// Close 关闭
func (mgr *Watcher) Close() {
	mgr.ClearWatch()
	_ = mgr.cli.Close()
	mgr.cli = nil
}

// watch 是内部实现的监视逻辑.
// ctx是上下文；key是需要监听的键（如果有）.
// prefix允许对具有匹配前缀的键进行操作。例如，“ Get（foo，WithPrefix（））” 可以返回“ foo1”，“ foo2”，依此类推.
// listener当监听到对应的事件时，将动作转发到Listener对应的实现.
func (mgr *Watcher) watch(ctx context.Context, key string, prefix bool, listener Listener) error {
	ctx1, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	var getResp *clientv3.GetResponse
	var err error
	if prefix {
		getResp, err = mgr.cli.Get(ctx1, key, clientv3.WithPrefix())
	} else {
		getResp, err = mgr.cli.Get(ctx1, key)
	}
	if err != nil {
		return err
	}

	leaderCtx := clientv3.WithRequireLeader(ctx)
	var watchChan clientv3.WatchChan
	if prefix {
		watchChan = mgr.cli.Watch(leaderCtx, key, clientv3.WithPrefix(), clientv3.WithRev(getResp.Header.Revision+1))
	} else {
		watchChan = mgr.cli.Watch(leaderCtx, key, clientv3.WithRev(getResp.Header.Revision+1))
	}
	for {
		select {
		case <-ctx.Done():
			err = errors.New("context canceled")
			goto EXIT
		case resp := <-watchChan:
			err = resp.Err()
			if err != nil {
				goto EXIT
			}
			if resp.Canceled {
				err = errors.New("watch failed and the stream was about to close")
				goto EXIT
			}
			for _, ev := range resp.Events {
				if ev.IsCreate() {
					listener.Create(ev.Kv.Key, ev.Kv.Value, ev.Kv.Version)
				} else if ev.IsModify() {
					listener.Modify(ev.Kv.Key, ev.Kv.Value, ev.Kv.Version)
				} else if ev.Type == mvccpb.DELETE {
					listener.Delete(ev.Kv.Key, ev.Kv.Version)
				} else {
				}
			}
		}
	}
EXIT:
	listener.Exit(err.Error())
	return nil
}

// Put 将一个键值对放入etcd中。
// 注意，key，value可以是纯字节数组，而string是该字节数组的不可变表示形式
func (mgr *Watcher) Put(ctx context.Context, key, value string) (err error) {
	newCtx, cancel := context.WithTimeout(ctx, timeOut)
	_, err = mgr.cli.Put(newCtx, key, value)
	cancel()
	if err != nil {
		return err
	}
	return nil
}
