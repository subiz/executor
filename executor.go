package executor

import (
	"sync"
	"time"
)

type Job struct {
	key  string
	data interface{}
}

// Executor executes job in parallel
type Executor struct {
	sync.Mutex
	workers map[string]*Worker
	handler func(string, interface{})
	njob    uint // total jobs
}

func New(f func(string, interface{})) *Executor {
	return &Executor{
		workers: make(map[string]*Worker),
		handler: f,
	}
}

// AddJob adds new job
// block if one of the queue is full
func (e *Executor) Add(key string, data interface{}) {
	e.Lock()
	defer e.Unlock()
	if _, found := e.workers[key]; !found {
		e.addWorker(key)
	}
	w := e.workers[key]
	e.njob++
	w.jobchan <- Job{key: key, data: data}
}

func (e *Executor) Stop() {
	for _, w := range e.workers {
		w.stop()
	}
}

// Info returns total jobs of each worker
func (e *Executor) Info() map[string]uint {
	e.Lock()
	defer e.Unlock()
	info := map[string]uint{}
	for id, w := range e.workers {
		info[id] = w.njob
	}
	return info
}

// Wait wait until all jobs is done
func (e *Executor) Wait() {
	for {
		njob, ndone := e.Count()
		if ndone == njob {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Count returns job count, done count
func (e *Executor) Count() (uint, uint) {
	e.Lock()
	defer e.Unlock()
	var ndone uint
	for _, w := range e.workers {
		ndone += w.ndone
	}
	return e.njob, ndone
}

// the caller must lock when to call this function
func (e *Executor) addWorker(id string) {
	w := NewWorker(id, e.handler, e)
	e.workers[id] = w
	go w.start()
}

// the caller must lock when to call this function
func (e *Executor) removeWorker(id string) {
	delete(e.workers, id)
}
