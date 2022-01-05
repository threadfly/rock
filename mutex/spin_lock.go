package mutex

import (
	"runtime"
	"sync/atomic"
)

type SpinLock struct {
	next        int64
	owner       int64
	enableSched bool
}

func NewSpinLock(enableSched bool) *SpinLock {
	return &SpinLock{
		enableSched: enableSched,
	}
}

func (s *SpinLock) Lock() {
	n := atomic.AddInt64(&s.next, 1)
	for (n - 1) != s.owner {
		if s.enableSched {
			runtime.Gosched()
		}
	}
}

func (s *SpinLock) Unlock() {
	atomic.AddInt64(&s.owner, 1)
}
