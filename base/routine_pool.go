package base

type Task interface {
	Run()
}

type RoutinePool struct {
	recv chan Task
}

func NewRoutinePool(cap uint16) *RoutinePool {
	pool := &RoutinePool{
		recv: make(chan Task, cap),
	}

	for i := 0; i < int(cap); i++ {
		go func(recv chan Task) {
			for {
				select {
				case t := <-recv:
					t.Run()
				}
			}
		}(pool.recv)
	}
	return pool
}

func (p *RoutinePool) Run(t Task) {
	p.recv <- t
}
