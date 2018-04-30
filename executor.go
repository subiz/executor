package executor

import (
	"errors"
	"sort"
)

type Job struct {
	Key  string
	Data interface{}
}

type Handler func(Job)

// Executor executes job in parallel
type Executor struct {
	// A pool of workers channels that are registered with the dispatcher
	workerPool map[int]*Worker

	maxWorkers     uint
	maxJobsInQueue uint // per worker
	handler        Handler
}

// maxJobsInQueue >= 2
func NewExecutor(maxWorkers, maxJobsInQueue uint, handler Handler) *Executor {
	if maxJobsInQueue < 2 {
		panic(errors.New("maxJobsInQueue must greater than 2"))
	}

	e := &Executor{
		workerPool:     map[int]*Worker{},
		maxWorkers:     maxWorkers,
		maxJobsInQueue: maxJobsInQueue,
		handler:        handler,
	}

	e.run()

	return e
}

// AddJob adds new job
// block if one of the queue is full
func (e *Executor) AddJob(job Job) {
	workerID := getWorkerID(job.Key, e.maxWorkers)
	worker := e.getWorker(workerID)

	// dispatch the job to the worker job channel
	worker.jobChannel <- job
}

func (e *Executor) run() {
	// Now, create all of our workers.
	for i := 1; i <= int(e.maxWorkers); i++ {
		workerID := i
		worker := NewWorker(uint(workerID), e.maxJobsInQueue, e.handler)
		go worker.start()

		e.workerPool[workerID] = worker
	}
}

func (e *Executor) Stop() {
	for _, worker := range e.workerPool {
		worker.stop()
	}
}

func (e *Executor) Info() map[int]uint {
	info := map[int]uint{}

	var keys []int
	for k := range e.workerPool {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		info[k] = e.workerPool[k].jobcount
	}

	return info
}

func (e *Executor) getWorker(id int) *Worker {
	return e.workerPool[id]
}
