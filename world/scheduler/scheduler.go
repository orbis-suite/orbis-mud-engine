package scheduler

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	NextRun time.Time
	RunFunc func() error
}

type Scheduler struct {
	mu   sync.Mutex
	jobs JobHeap
	wake chan struct{}
	quit chan struct{}
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		jobs: make(JobHeap, 0),
		wake: make(chan struct{}, 1),
		quit: make(chan struct{}),
	}
	heap.Init(&s.jobs)
	go s.run()
	return s
}

func (s *Scheduler) Add(job *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	heap.Push(&s.jobs, job)

	select {
	case s.wake <- struct{}{}:
	default:
	} // wake the loop upon jobs updating
}

func (s *Scheduler) run() {
	for {
		s.mu.Lock()
		if len(s.jobs) == 0 {
			s.mu.Unlock()
			select {
			case <-s.wake:
				continue
			case <-s.quit:
				return
			}
		}

		next := s.jobs[0]
		now := time.Now()
		wait := next.NextRun.Sub(now)
		s.mu.Unlock()

		if wait > 0 {
			select {
			case <-time.After(wait):
			case <-s.wake:
				continue
			case <-s.quit:
				return
			}
		}

		s.mu.Lock()
		next = heap.Pop(&s.jobs).(*Job)
		s.mu.Unlock()

		err := next.RunFunc()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.quit)
}
