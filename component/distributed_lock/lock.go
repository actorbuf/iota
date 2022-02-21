package distributed_lock

import "sync"

type Lock struct {
	Driver sync.Locker
}
