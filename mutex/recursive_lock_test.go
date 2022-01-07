package mutex

import (
	"sync"
	"testing"
	"time"
)

func doRecursiveLock(l *RecursiveLock) {
	l.Lock()
	l.Lock()
	l.Lock()
	count++
	l.Unlock()
	l.Unlock()
	l.Unlock()
}

func TestRecursiveLock(t *testing.T) {
	//mx := NewRecursiveLock(&sync.Mutex{})
	mx := NewSpinLock(true)
	//mx := &sync.Mutex{}
	l := NewRecursiveLock(mx)
	var (
		wg    sync.WaitGroup
		begin = time.Now()
	)
	wg.Add(loop)
	for i := 0; i < loop; i++ {
		go func(no int) {
			t.Logf("go no.%d", no)
			doRecursiveLock(l)
			wg.Done()
		}(i)
	}
	wg.Wait()
	if count != loop {
		t.Fatalf("count:%d loop:%d", count, loop)
	} else {
		t.Logf("count:%d loop:%d, cost:%s", count, loop, time.Now().Sub(begin))
	}
}
