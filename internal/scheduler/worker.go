package scheduler

import (
	"context"
	"log"
	"sync"
	"time"
)

type Worker struct {
	id      int
	Current *Job
	jobChan chan *Job
	mutex   sync.Mutex
}

func NewWorker(id int) *Worker {
	return &Worker{
		id:      id,
		jobChan: make(chan *Job, 1),
		Current: nil,
		mutex:   sync.Mutex{},
	}
}

func (w *Worker) Start(ctx context.Context) {

	log.Printf("Worker %d started", w.id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d shutting down", w.id)
			return

		case job, ok := <-w.jobChan:
			if !ok {
				log.Printf("Worker %d: channel closed", w.id)
				return
			}

			if job != nil {
				w.process(ctx, job)
			}
		}
	}
}

func (w *Worker) process(ctx context.Context, job *Job) {

	w.mutex.Lock()
	w.Current = job
	w.mutex.Unlock()

	log.Printf("Worker %d processing job %s (priority: %d)",
		w.id, job.Name, job.Priority)

	start := time.Now()
	job.Execute(ctx, job.Duration)

	log.Printf("Worker %d completed job %s in %v",
		w.id, job.Name, time.Since(start))

	w.mutex.Lock()
	w.Current = nil
	w.mutex.Unlock()

}

func (w *Worker) IsFree() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.Current == nil
}

func (w *Worker) GetCurrent() *Job {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.Current
}
