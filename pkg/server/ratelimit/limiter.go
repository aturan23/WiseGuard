package ratelimit

import (
	"net"
	"sync"
	"time"
)

type IPRateLimiter struct {
	mu     sync.RWMutex
	limits map[string]bucket
	rate   int
	burst  int
	ttl    time.Duration
}

type bucket struct {
	tokens    int
	lastReset time.Time
	blocked   bool
	attempts  int
}

func NewIPRateLimiter(rate, burst int, ttl time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		limits: make(map[string]bucket),
		rate:   rate,
		burst:  burst,
		ttl:    ttl,
	}

	go limiter.periodicCleanup()

	return limiter
}

func (rl *IPRateLimiter) AllowConnection(addr net.Addr) bool {
	ip := extractIP(addr)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.limits[ip]
	if !exists {
		b = bucket{
			tokens:    rl.burst,
			lastReset: time.Now(),
		}
	}

	// Check if IP is blocked
	if b.blocked {
		rl.limits[ip] = b
		return false
	}

	// Reset tokens if needed
	if time.Since(b.lastReset) > time.Minute {
		b.tokens = rl.burst
		b.lastReset = time.Now()
	}

	// Check if we have tokens
	if b.tokens > 0 {
		b.tokens--
		rl.limits[ip] = b
		return true
	}

	// Increment failed attempts
	b.attempts++

	// Block IP if too many failed attempts
	if b.attempts > rl.burst*2 {
		b.blocked = true
	}

	rl.limits[ip] = b
	return false
}

func (rl *IPRateLimiter) periodicCleanup() {
	ticker := time.NewTicker(rl.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

func (rl *IPRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, bucket := range rl.limits {
		if now.Sub(bucket.lastReset) > rl.ttl {
			delete(rl.limits, ip)
		}
	}
}

func extractIP(addr net.Addr) string {
	ip, _, _ := net.SplitHostPort(addr.String())
	return ip
}
