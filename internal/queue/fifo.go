package queue

import "sync"

type Fifo[T any] struct {
	cond    *sync.Cond
	queue   []T
	stopped bool
}

func NewFifo[T any]() *Fifo[T] {
	mu := sync.Mutex{}
	cond := sync.NewCond(&mu)

	return &Fifo[T]{
		cond:  cond,
		queue: make([]T, 0),
	}
}

func (f *Fifo[T]) Shutdown() {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	f.stopped = true
	f.cond.Broadcast()
}

func (f *Fifo[T]) Add(v T) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	if f.stopped {
		return
	}

	f.queue = append(f.queue, v)
	f.cond.Signal()
}

func (f *Fifo[T]) Get() (T, bool) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	if len(f.queue) == 0 && !f.stopped {
		f.cond.Wait()
	}
	if len(f.queue) == 0 {
		var zero T
		return zero, f.stopped
	}
	val := f.queue[0]
	f.queue = f.queue[1:]
	return val, f.stopped
}
