package ratelimit

import (
	"net"
	"sync"
	"time"
)

type ConnectionLimiter struct {
	mu          sync.RWMutex
	connections map[string]connStats
	maxPerIP    int
	ttl         time.Duration
}

type connStats struct {
	count     int
	lastReset time.Time
}

func NewConnectionLimiter(maxPerIP int, ttl time.Duration) *ConnectionLimiter {
	limiter := &ConnectionLimiter{
		connections: make(map[string]connStats),
		maxPerIP:    maxPerIP,
		ttl:         ttl,
	}
	go limiter.periodicCleanup()
	return limiter
}

func (cl *ConnectionLimiter) AllowConnection(addr net.Addr) bool {
	ip := extractIP(addr)

	cl.mu.Lock()
	defer cl.mu.Unlock()

	stats, exists := cl.connections[ip]
	if !exists {
		stats = connStats{
			lastReset: time.Now(),
		}
	}

	// Reset counter if TTL expired
	if time.Since(stats.lastReset) > cl.ttl {
		stats.count = 0
		stats.lastReset = time.Now()
	}

	if stats.count >= cl.maxPerIP {
		return false
	}

	stats.count++
	cl.connections[ip] = stats
	return true
}

func (cl *ConnectionLimiter) RemoveConnection(addr net.Addr) {
	ip := extractIP(addr)

	cl.mu.Lock()
	defer cl.mu.Unlock()

	if stats, exists := cl.connections[ip]; exists {
		stats.count--
		if stats.count <= 0 {
			delete(cl.connections, ip)
		} else {
			cl.connections[ip] = stats
		}
	}
}

func (cl *ConnectionLimiter) periodicCleanup() {
	ticker := time.NewTicker(cl.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		cl.cleanup()
	}
}

func (cl *ConnectionLimiter) cleanup() {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	now := time.Now()
	for ip, stats := range cl.connections {
		if now.Sub(stats.lastReset) > cl.ttl {
			delete(cl.connections, ip)
		}
	}
}
