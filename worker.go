package executor

type Worker struct {
	id         uint
	jobChannel chan Job
	quit       chan bool
	handler    Handler
	jobcount   uint
}

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func NewWorker(id, maxJobs uint, handler Handler) *Worker {
	// Create, and return the worker.
	return &Worker{
		id:         id,
		jobChannel: make(chan Job, maxJobs-1),
		quit:       make(chan bool),
		handler:    handler,
	}
}

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) start() {
	for {
		select {
		case job := <-w.jobChannel:
			w.jobcount++
			w.handler(job)
		case <-w.quit:
			return
		}
	}
}

// Stop tells the worker to stop listening for work requests.
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) stop() {
	w.quit <- true
}
