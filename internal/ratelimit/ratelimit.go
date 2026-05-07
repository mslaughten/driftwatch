// Package ratelimit provides a simple token-bucket rate limiter
// to prevent alert flooding when many config files change at once.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a token-bucket rate limiter.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
	clock    func() time.Time
}

// New creates a Limiter that allows up to max events with a refill
// rate of ratePerSec tokens per second.
func New(max float64, ratePerSec float64) *Limiter {
	return &Limiter{
		tokens:   max,
		max:      max,
		rate:     ratePerSec,
		lastTick: time.Now(),
		clock:    time.Now,
	}
}

// Allow reports whether an event may proceed. It consumes one token
// if available, refilling the bucket based on elapsed time first.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock()
	elapsed := now.Sub(l.lastTick).Seconds()
	l.lastTick = now

	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}

	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}

// Tokens returns the current number of available tokens (for inspection/testing).
func (l *Limiter) Tokens() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.tokens
}
