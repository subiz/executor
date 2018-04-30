package executor

type Job struct {
	Key  string
	Data interface{}
}

type Handler func(Job)

// Executor executes job in parallel
type Executor struct {
	workers        []*Worker
	maxWorkers     uint
	maxJobsInQueue uint // per worker
	handler        Handler
}

// maxJobsInQueue >= 2
func NewExecutor(maxWorkers, maxJobsInQueue uint, handler Handler) *Executor {
	if maxJobsInQueue < 2 {
		panic("maxJobsInQueue must greater than 2")
	}

	e := &Executor{
		workers:        make([]*Worker, 0, maxWorkers),
		maxWorkers:     maxWorkers,
		maxJobsInQueue: maxJobsInQueue,
		handler:        handler,
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
func (e *Executor) AddJob(job Job) {
	worker := e.workers[getWorkerID(job.Key, e.maxWorkers)]
	worker.jobChannel <- job
}

func (e *Executor) Stop() {
	for _, worker := range e.workers {
		worker.stop()
	}
}

func (e *Executor) Info() map[int]uint {
	info := map[int]uint{}

	for i, w := range e.workers {
		info[i] = w.jobcount
	}

	return info
}
