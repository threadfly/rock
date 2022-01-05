package base

import (
	"sync"
	"testing"
)

type task struct {
	t    *testing.T
	num  int
	recv chan int
}

func Newtask(t *testing.T, num int) *task {
	return &task{
		t:    t,
		num:  num,
		recv: make(chan int, 1),
	}
}

func (t *task) Run() {
	t.t.Logf("num:%d", t.num)
	t.recv <- t.num
	close(t.recv)
}

func (t *task) Wait() int {
	return <-t.recv
}

func TestRoutine(t *testing.T) {
	pool := NewRoutinePool(10)
	wg := new(sync.WaitGroup)
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func(num int) {
			task := Newtask(t, num)
			pool.Run(task)
			task.Wait()
			wg.Done()
		}(i)
	}
	wg.Wait()
}
