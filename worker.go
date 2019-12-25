package executor

import "time"

const TIMEOUT = 3 * time.Second

type Worker struct {
	id      string
	jobchan chan Job
	quit    chan bool
	handler func(string, interface{})
	njob    uint // total jobs
	ndone   uint // total done jobs
	mgr     *Executor
}

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func NewWorker(id string, handler func(string, interface{}), mgr *Executor) *Worker {
	return &Worker{
		id:      id,
		jobchan: make(chan Job, 20),
		quit:    make(chan bool),
		handler: handler,
		mgr:     mgr,
	}
}

// start runs the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) start() {
loop:
	for {
		select {
		case job := <-w.jobchan:
			w.njob++
			w.handler(job.key, job.data)
			w.ndone++
		case <-w.quit:
			break loop
		case <-time.After(TIMEOUT):
			w.mgr.Lock()
			if len(w.jobchan) == 0 {
				w.mgr.removeWorker(w.id)
				w.mgr.Unlock()
				break loop
			}
			w.mgr.Unlock()
		}
	}
}

// stop tells the worker to stop listening for work requests.
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) stop() { w.quit <- true }
