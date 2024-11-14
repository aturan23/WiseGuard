package protection

import (
	"sync"
	"time"
)

type TokenBucket struct {
	tokens          map[string]*bucket
	capacity        int64
	fillRate        float64
	stopCh          chan struct{}
	cleanupInterval time.Duration
	mu              sync.RWMutex
}

type bucket struct {
	tokens     int64
	lastUpdate time.Time
}

func NewTokenBucket(capacity int64, fillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:          make(map[string]*bucket),
		capacity:        capacity,
		fillRate:        fillRate,
		stopCh:          make(chan struct{}),
		cleanupInterval: time.Minute * 5, // Cleanup unused buckets every 5 minutes
	}
}

// Start begins the token bucket cleaner and refill processes
func (t *TokenBucket) Start() {
	// Start the cleanup routine
	go t.cleanupRoutine()

	// Start the refill routine
	go t.refillRoutine()
}

// Stop gracefully stops the token bucket processes
func (t *TokenBucket) Stop() {
	close(t.stopCh)
}

func (t *TokenBucket) cleanupRoutine() {
	ticker := time.NewTicker(t.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.cleanup()
		case <-t.stopCh:
			return
		}
	}
}

func (t *TokenBucket) refillRoutine() {
	// Refill tokens more frequently than cleanup
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.refillAll()
		case <-t.stopCh:
			return
		}
	}
}

func (t *TokenBucket) cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	// Remove buckets that haven't been used for a while
	for ip, bucket := range t.tokens {
		if now.Sub(bucket.lastUpdate) > t.cleanupInterval {
			delete(t.tokens, ip)
		}
	}
}

func (t *TokenBucket) refillAll() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for _, b := range t.tokens {
		t.refillBucket(b, now)
	}
}

func (t *TokenBucket) refillBucket(b *bucket, now time.Time) {
	elapsed := now.Sub(b.lastUpdate).Seconds()
	newTokens := int64(elapsed * t.fillRate)

	if newTokens > 0 {
		b.tokens = min(b.tokens+newTokens, t.capacity)
		b.lastUpdate = now
	}
}

func (t *TokenBucket) Take(ip string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	b, exists := t.tokens[ip]
	if !exists {
		b = &bucket{
			tokens:     t.capacity,
			lastUpdate: time.Now(),
		}
		t.tokens[ip] = b
	}

	// Refill the bucket
	t.refillBucket(b, time.Now())

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// GetTokens returns the current number of tokens for an IP
func (t *TokenBucket) GetTokens(ip string) int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if b, exists := t.tokens[ip]; exists {
		t.refillBucket(b, time.Now())
		return b.tokens
	}

	return t.capacity
}
