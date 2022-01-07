package mutex

import (
	"reflect"
	"rock/base"
	"sync"
	"sync/atomic"
)

type RecursiveLock struct {
	sync.Locker
	gid int64
	cnt int
}

func NewRecursiveLock(l sync.Locker) *RecursiveLock {
	r := &RecursiveLock{
		Locker: l,
		gid:    -1,
	}
	if reflect.TypeOf(r).String() == reflect.TypeOf(l).String() {
		panic("cannot use recursive lock as inner lock")
	}
	return r
}

func (rl *RecursiveLock) Lock() {
	gid := (int64)(base.GoID())
	if rl.gid != gid {
		rl.Locker.Lock()
		if !atomic.CompareAndSwapInt64(&rl.gid, -1, gid) {
			panic("lock unreachable")
		}
	}
	rl.cnt++
}

func (rl *RecursiveLock) Unlock() {
	gid := (int64)(base.GoID())
	if rl.gid != gid {
		panic("lock before unlock")
	}

	rl.cnt--
	if rl.cnt == 0 {
		atomic.CompareAndSwapInt64(&rl.gid, gid, -1)
		rl.Locker.Unlock()
	}
}
