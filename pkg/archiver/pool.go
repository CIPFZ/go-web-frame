package archiver

import (
	"runtime"
	"sync"
)

// Job represents a function to be executed by a worker.
type Job func()

// Pool is a worker pool that manages a number of concurrent goroutines.
type Pool struct {
	numWorkers int
	jobs       chan Job
	wg         sync.WaitGroup
	stopOnce   sync.Once
	stopped    chan struct{}
}

// NewPool creates a new worker pool with a specified number of workers.
// If numWorkers is less than 1, it defaults to the number of available CPU cores.
func NewPool(numWorkers int) *Pool {
	if numWorkers < 1 {
		numWorkers = runtime.NumCPU()
	}
	return &Pool{
		numWorkers: numWorkers,
		// Use a buffered channel to prevent blocking on submit if workers are busy
		jobs:    make(chan Job, numWorkers),
		stopped: make(chan struct{}),
	}
}

// Start spawns the worker goroutines.
func (p *Pool) Start() {
	for i := 0; i < p.numWorkers; i++ {
		go p.worker()
	}
}

// worker is the heart of the pool, listening for jobs or shutdown signals.
func (p *Pool) worker() {
	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				// 如果 channel 已关闭 (ok == false)，说明要停止了，直接返回
				return
			}
			job()
			p.wg.Done()
		case <-p.stopped:
			return
		}
	}
}

// Submit adds a job to the pool for execution.
// It will panic if the pool has been stopped.
func (p *Pool) Submit(job Job) {
	select {
	case <-p.stopped:
		panic("archiver: submitting job to a stopped pool")
	default:
	}

	p.wg.Add(1)
	p.jobs <- job
}

// Wait blocks until all submitted jobs have completed.
func (p *Pool) Wait() {
	p.wg.Wait()
}

// Stop gracefully shuts down all workers.
// It ensures that all previously submitted jobs are completed before stopping.
func (p *Pool) Stop() {
	p.stopOnce.Do(func() {
		// Wait for any pending jobs to finish
		p.Wait()
		// Signal workers to stop
		close(p.stopped)
		// Close the jobs channel
		close(p.jobs)
	})
}
