package scheduler

import (
	"container/heap"
	"context"
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	ctx     context.Context
	cancel  context.CancelFunc
	jobs    JobPriorityQueue
	jobID   int64
	workers []*Worker
	mutex   sync.Mutex
	wg      sync.WaitGroup
}

func New(numWorkers int, parent context.Context) *Scheduler {
	ctx, cancel := context.WithCancel(parent)
	s := &Scheduler{
		jobs:    make(JobPriorityQueue, 0),
		workers: make([]*Worker, numWorkers),
		ctx:     ctx,
		cancel:  cancel,
		mutex:   sync.Mutex{},
	}
	heap.Init(&s.jobs)

	for i := 0; i < numWorkers; i++ {
		s.workers[i] = NewWorker(i)
		s.wg.Add(1)
		go func(w *Worker) {
			defer s.wg.Done()
			w.Start(ctx)
		}(s.workers[i])
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.dispatch()
	}()

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
	s.mutex.Lock()
	if s.jobs.Len() == 0 {
		s.mutex.Unlock()
		return
	}
	job := heap.Pop(&s.jobs).(*Job)
	s.mutex.Unlock()

	for _, worker := range s.workers {
		select {
		case worker.jobChan <- job:
			log.Printf("Job %s dispatched to worker %d", job.Name, worker.id)
			return
		default:
		}
	}

	s.mutex.Lock()
	heap.Push(&s.jobs, job)
	s.mutex.Unlock()
}

func (s *Scheduler) Schedule(job *Job) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.jobID++
	job.ID = s.jobID
	heap.Push(&s.jobs, job)
}

func (s *Scheduler) Shutdown(ctx context.Context) error {
	log.Println("Shutting down scheduler...")
	s.cancel()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Scheduler shutdown finished successfully")
		return nil
	case <-ctx.Done():
		log.Println("Scheduler shutdown timeout")
		return ctx.Err()
	}
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
	s.mutex.Lock()
	defer s.mutex.Unlock()
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
