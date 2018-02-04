package controller

import "sync"

const workQueueMax int = 100

// asyncRunner is a runner that will do work
// against a single id serially and do work against
// different ids in parallel
type asyncRunner struct {
	lock       *sync.Mutex
	workQueues map[string]workQueue
}

type workQueue struct {
	lock  *sync.Mutex
	queue chan func()
}

func newAsyncRunner() *asyncRunner {
	return &asyncRunner{
		lock:       &sync.Mutex{},
		workQueues: map[string]workQueue{},
	}
}

func newWorkQueue() workQueue {
	return workQueue{
		lock:  &sync.Mutex{},
		queue: make(chan func(), workQueueMax),
	}
}

func (r *asyncRunner) getOrCreateWorkQueue(id string) workQueue {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.workQueues[id]; !ok {
		r.workQueues[id] = newWorkQueue()
	}

	return r.workQueues[id]
}

func (r *asyncRunner) run(id string, f func()) {
	// Add to end of queue
	wq := r.getOrCreateWorkQueue(id)
	wq.queue <- f

	// Do one item of work from top of queue
	// to serialize work per id.
	wq.lock.Lock()
	defer wq.lock.Unlock()

	w := <-wq.queue
	w()
}

func (r *asyncRunner) runAsync(id string, f func()) {
	go r.run(id, f)
}
