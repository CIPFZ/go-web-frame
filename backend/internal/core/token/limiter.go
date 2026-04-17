package token

import "sync"

type InMemoryLimiter struct {
	mu     sync.Mutex
	counts map[uint]int
}

func NewInMemoryLimiter() *InMemoryLimiter {
	return &InMemoryLimiter{
		counts: make(map[uint]int),
	}
}

func (l *InMemoryLimiter) Acquire(tokenID uint, maxConcurrency int) bool {
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.counts[tokenID] >= maxConcurrency {
		return false
	}
	l.counts[tokenID]++
	return true
}

func (l *InMemoryLimiter) Release(tokenID uint) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.counts[tokenID] <= 1 {
		delete(l.counts, tokenID)
		return
	}
	l.counts[tokenID]--
}
