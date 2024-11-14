package protection

import (
	"net"
	"sync"
	"time"
)

type FailedAttempts struct {
	attempts        map[string]*attemptInfo
	maxAttempts     int
	blockTime       time.Duration
	cleanupInterval time.Duration
	mu              sync.RWMutex
	stopCh          chan struct{}
}

type attemptInfo struct {
	count     int
	lastFail  time.Time
	blockedAt time.Time
}

func NewFailedAttempts(maxAttempts int, blockTime time.Duration) *FailedAttempts {
	fa := &FailedAttempts{
		attempts:        make(map[string]*attemptInfo),
		maxAttempts:     maxAttempts,
		blockTime:       blockTime,
		cleanupInterval: time.Minute * 5, // Cleanup every 5 minutes
		stopCh:          make(chan struct{}),
	}

	// Start cleanup goroutine
	go fa.cleanup()

	return fa
}

func (f *FailedAttempts) RegisterFailure(addr net.Addr) {
	f.mu.Lock()
	defer f.mu.Unlock()

	ip := addr.String()
	info, exists := f.attempts[ip]

	if !exists {
		info = &attemptInfo{}
		f.attempts[ip] = info
	}

	info.count++
	info.lastFail = time.Now()

	if info.count >= f.maxAttempts {
		info.blockedAt = time.Now()
	}
}

func (f *FailedAttempts) IsBlocked(addr net.Addr) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	ip := addr.String()
	info, exists := f.attempts[ip]
	if !exists {
		return false
	}

	// If blocked and block time hasn't expired
	if info.count >= f.maxAttempts && time.Since(info.blockedAt) < f.blockTime {
		return true
	}

	// If the last failure was long ago, reset the counter
	if time.Since(info.lastFail) > f.blockTime {
		delete(f.attempts, ip)
		return false
	}

	return false
}

func (f *FailedAttempts) ResetFailures(addr net.Addr) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.attempts, addr.String())
}

func (f *FailedAttempts) cleanup() {
	ticker := time.NewTicker(f.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f.mu.Lock()
			now := time.Now()

			// Remove expired entries
			for ip, info := range f.attempts {
				// Remove if block time expired or last failure was too long ago
				if (info.count >= f.maxAttempts && now.Sub(info.blockedAt) > f.blockTime) ||
					now.Sub(info.lastFail) > f.blockTime {
					delete(f.attempts, ip)
				}
			}
			f.mu.Unlock()

		case <-f.stopCh:
			return
		}
	}
}

func (f *FailedAttempts) Stop() {
	close(f.stopCh)
}

// GetFailureCount returns current failure count for an IP
func (f *FailedAttempts) GetFailureCount(addr net.Addr) int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if info, exists := f.attempts[addr.String()]; exists {
		return info.count
	}
	return 0
}

// GetBlockTimeRemaining returns remaining block time for an IP
func (f *FailedAttempts) GetBlockTimeRemaining(addr net.Addr) time.Duration {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if info, exists := f.attempts[addr.String()]; exists && info.count >= f.maxAttempts {
		remaining := f.blockTime - time.Since(info.blockedAt)
		if remaining > 0 {
			return remaining
		}
	}
	return 0
}
