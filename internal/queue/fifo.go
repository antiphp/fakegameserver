// Package queue provides queue implementations.
package queue

import "sync"

// Fifo is a first-in-first-out queue.
type Fifo[T any] struct {
	cond    *sync.Cond
	queue   []T
	stopped bool
}

// NewFifo creates a new Fifo queue.
func NewFifo[T any]() *Fifo[T] {
	mu := sync.Mutex{}
	cond := sync.NewCond(&mu)

	return &Fifo[T]{
		cond:  cond,
		queue: make([]T, 0),
	}
}

// Shutdown stops the queue and wakes up all waiting goroutines.
func (f *Fifo[T]) Shutdown() {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	f.stopped = true
	f.cond.Broadcast()
}

// Add adds an item to the queue and wakes up one waiting goroutine.
func (f *Fifo[T]) Add(v T) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	if f.stopped {
		return
	}

	f.queue = append(f.queue, v)
	f.cond.Signal()
}

// Get retrieves an item from the queue.
//
// If the queue is empty, it waits until an item is added or the queue is stopped.
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
