package scheduler

import (
	"context"
	"log"
	"time"
)

type Worker struct {
	id      int
	Current *Job
	jobChan chan *Job
}

func NewWorker(id int) *Worker {
	return &Worker{
		id:      id,
		jobChan: make(chan *Job, 1),
		Current: nil,
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

	w.Current = job

	log.Printf("Worker %d processing job %s (priority: %d)",
		w.id, job.Name, job.Priority)

	start := time.Now()
	job.Execute(ctx, job.Duration)

	log.Printf("Worker %d completed job %s in %v",
		w.id, job.Name, time.Since(start))

	w.Current = nil

}

func (w *Worker) IsFree() bool {

	return w.Current == nil
}

func (w *Worker) GetCurrent() *Job {
	return w.Current
}
