package executor

import "time"

type Job struct {
	key  string
	data interface{}
}

// Executor executes job in parallel
type Executor struct {
	workers        []*Worker
	maxWorkers     uint
	maxJobsInQueue uint // per worker
	handler        func(string, interface{})
	jobCount       uint
}

func New(nworkers uint, f func(string, interface{})) *Executor {
	e := &Executor{
		workers:        make([]*Worker, 0, nworkers),
		maxWorkers:     nworkers,
		maxJobsInQueue: 20,
		handler:        f,
	}

	// creates and runs workers
	for i := uint(0); i < e.maxWorkers; i++ {
		worker := NewWorker(i, e.maxJobsInQueue, e.handler)
		e.workers = append(e.workers, worker)
		go worker.start()
	}

	return e
}

// AddJob adds new job
// block if one of the queue is full
func (e *Executor) Add(key string, data interface{}) {
	e.jobCount++
	worker := e.workers[getWorkerID(key, e.maxWorkers)]
	worker.jobChannel <- Job{key: key, data: data}
}

func (e *Executor) Stop() {
	for _, worker := range e.workers {
		worker.stop()
	}
}

func (e *Executor) Info() map[int]uint {
	info := map[int]uint{}

	for i, w := range e.workers {
		info[i] = w.jobCount
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

// Count return job count, done count
func (e *Executor) Count() (uint, uint) {
	var doneCount uint
	for _, w := range e.workers {
		doneCount += w.doneCount
	}
	return e.jobCount, doneCount
}
