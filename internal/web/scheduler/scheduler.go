package scheduler

import (
	"container/heap"
	"context"
	"log"
	"time"
)

type Scheduler struct {
	jobs    JobPriorityQueue
	jobID   int64
	workers []*Worker
	ctx     context.Context
	cancel  context.CancelFunc
}

func New(numWorkers int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Scheduler{
		jobs:    make(JobPriorityQueue, 0),
		workers: make([]*Worker, numWorkers),
		ctx:     ctx,
		cancel:  cancel,
	}

	heap.Init(&s.jobs)

	for i := 0; i < numWorkers; i++ {
		s.workers[i] = NewWorker(i)

		go s.workers[i].Start(ctx)
	}

	go s.dispatch()

	return s
}

func (s *Scheduler) dispatch() {

	log.Println("Dispatcher started")

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			log.Println("Dispatcher shutting down")
			return

		case <-ticker.C:
			s.tryDispatch()
		}
	}
}

func (s *Scheduler) tryDispatch() {

	if s.jobs.Len() == 0 {
		return
	}

	for _, worker := range s.workers {
		if worker.IsFree() && s.jobs.Len() > 0 {
			job := heap.Pop(&s.jobs).(*Job)

			select {
			case worker.jobChan <- job:
				log.Printf("Job %s dispatched to worker %d", job.Name, worker.id)
			default:
				heap.Push(&s.jobs, job)
			}
		}
	}
}

func (s *Scheduler) Schedule(job *Job) {

	s.jobID++
	job.ID = s.jobID
	heap.Push(&s.jobs, job)

	log.Printf("Job %s scheduled (priority: %d)", job.Name, job.Priority)
}

func (s *Scheduler) Shutdown() {
	log.Println("Shutting down scheduler...")
	s.cancel()

	for _, w := range s.workers {
		close(w.jobChan)
	}

	log.Println("All workers stopped")
}

func (s *Scheduler) GetStatus() []WorkerStatus {
	statuses := make([]WorkerStatus, len(s.workers))

	for i, worker := range s.workers {
		current := worker.GetCurrent()
		statuses[i] = WorkerStatus{
			WorkerID:   worker.id,
			IsBusy:     current != nil,
			CurrentJob: current,
		}
	}

	return statuses
}

func (s *Scheduler) GetQueuedJobs() []*Job {

	jobs := make([]*Job, s.jobs.Len())
	for i := 0; i < len(jobs); i++ {
		jobs[i] = s.jobs[i]
	}

	return jobs
}

type WorkerStatus struct {
	WorkerID   int  `json:"worker_id"`
	IsBusy     bool `json:"is_busy"`
	CurrentJob *Job `json:"current_job,omitempty"`
}
