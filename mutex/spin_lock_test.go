package mutex

import (
	"sync"
	"testing"
	"time"
)

var (
	count int
	loop  int = 100
)

func doSpinLock(l *SpinLock) {
	l.Lock()
	defer l.Unlock()
	count++
}

func TestSpinLock(t *testing.T) {
	l := NewSpinLock(false)
	var (
		wg    sync.WaitGroup
		begin = time.Now()
	)
	wg.Add(loop)
	for i := 0; i < loop; i++ {
		go func(no int) {
			t.Logf("go no.%d", no)
			doSpinLock(l)
			wg.Done()
		}(i)
	}
	wg.Wait()
	if count != loop {
		t.Fatalf("count:%d loop:%d", count, loop)
	} else {
		t.Logf("count:%d loop:%d, cost:%s", count, loop, time.Now().Sub(begin))
	}

	count = 0
	l = NewSpinLock(true)
	begin = time.Now()
	wg.Add(loop)
	for i := 0; i < loop; i++ {
		go func() {
			doSpinLock(l)
			wg.Done()
		}()
	}
	wg.Wait()
	if count != loop {
		t.Fatalf("count:%d loop:%d", count, loop)
	} else {
		t.Logf("count:%d loop:%d, cost:%s", count, loop, time.Now().Sub(begin))
	}
}
