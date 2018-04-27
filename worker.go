package executor

import (
	"fmt"
)

type Counter struct {
	Total uint
	Done  uint
}

type Worker struct {
	id         uint
	jobChannel chan Job
	quit       chan bool
	handler    Handler
	counter    Counter
}

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func NewWorker(id, maxJobs uint, handler Handler) *Worker {
	// Create, and return the worker.
	w := &Worker{
		id:         id,
		jobChannel: make(chan Job, maxJobs),
		quit:       make(chan bool),
		handler:    handler,
		counter:    Counter{},
	}

	return w
}

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) start() {
	for {
		select {
		case job := <-w.jobChannel:
			// Receive a job.
			fmt.Printf("Worker %d: Received job\n", w.id)
			err := w.handler(job)

			if err != nil {
				fmt.Printf("Worker %d: error process job - %s\n", w.id, err)
			}

			w.counter.Done++
		case <-w.quit:
			// We have been asked to stop.
			fmt.Printf("Worker %d: stopping\n", w.id)
			return
		}
	}
}

// Stop tells the worker to stop listening for work requests.
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) stop() {
	w.quit <- true
}
