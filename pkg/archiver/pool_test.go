package archiver

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	t.Run("PositiveWorkers", func(t *testing.T) {
		p := NewPool(4)
		if p.numWorkers != 4 {
			t.Errorf("expected 4 workers, got %d", p.numWorkers)
		}
	})

	t.Run("ZeroWorkers", func(t *testing.T) {
		p := NewPool(0)
		expected := runtime.NumCPU()
		if p.numWorkers != expected {
			t.Errorf("expected %d workers, got %d", expected, p.numWorkers)
		}
	})

	t.Run("NegativeWorkers", func(t *testing.T) {
		p := NewPool(-1)
		expected := runtime.NumCPU()
		if p.numWorkers != expected {
			t.Errorf("expected %d workers, got %d", expected, p.numWorkers)
		}
	})
}

func TestPoolExecution(t *testing.T) {
	var counter int64
	numJobs := 100
	numWorkers := 5

	pool := NewPool(numWorkers)
	pool.Start()

	for i := 0; i < numJobs; i++ {
		pool.Submit(func() {
			atomic.AddInt64(&counter, 1)
			time.Sleep(1 * time.Millisecond) // Simulate work
		})
	}

	pool.Wait()
	pool.Stop()

	if atomic.LoadInt64(&counter) != int64(numJobs) {
		t.Errorf("expected counter to be %d, got %d", numJobs, counter)
	}
}

func TestPoolStop(t *testing.T) {
	pool := NewPool(2)
	pool.Start()

	var wg sync.WaitGroup
	wg.Add(1)
	pool.Submit(func() {
		// This job will run
		time.Sleep(5 * time.Millisecond)
		wg.Done()
	})

	wg.Wait() // Ensure the first job is processed
	pool.Stop()

	// Test that submitting to a stopped pool panics
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when submitting to a stopped pool, but got none")
		}
	}()
	pool.Submit(func() {
		t.Error("this job should not run")
	})
}

func TestPoolReuse(t *testing.T) {
	var counter int64
	numJobs := 10
	numWorkers := 2

	pool := NewPool(numWorkers)
	pool.Start()

	// First batch
	for i := 0; i < numJobs; i++ {
		pool.Submit(func() {
			atomic.AddInt64(&counter, 1)
		})
	}
	pool.Wait()
	if atomic.LoadInt64(&counter) != int64(numJobs) {
		t.Errorf("expected counter to be %d after first batch, got %d", numJobs, counter)
	}

	// Second batch
	for i := 0; i < numJobs; i++ {
		pool.Submit(func() {
			atomic.AddInt64(&counter, 1)
		})
	}
	pool.Wait()
	if atomic.LoadInt64(&counter) != int64(numJobs*2) {
		t.Errorf("expected counter to be %d after second batch, got %d", numJobs*2, counter)
	}

	pool.Stop()
}
