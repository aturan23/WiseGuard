package protection

import (
	"errors"
	"net"
	"sync"
	"time"
)

// ProtectionConfig holds all protection-related configuration
type ProtectionConfig struct {
	// Slow Request Protection
	MinReadRate int64         `env:"PROTECTION_MIN_READ_RATE" envDefault:"100"` // bytes per second
	ReadTimeout time.Duration `env:"PROTECTION_READ_TIMEOUT" envDefault:"10s"`

	// IP Filtering
	EnableIPFilter bool     `env:"PROTECTION_ENABLE_IP_FILTER" envDefault:"true"`
	IPWhitelist    []string `env:"PROTECTION_IP_WHITELIST" envDefault:""`
	IPBlacklist    []string `env:"PROTECTION_IP_BLACKLIST" envDefault:""`

	// Failed Attempts
	MaxFailedAttempts int           `env:"PROTECTION_MAX_FAILED_ATTEMPTS" envDefault:"5"`
	FailedBlockTime   time.Duration `env:"PROTECTION_FAILED_BLOCK_TIME" envDefault:"15m"`

	// Memory Protection
	MemoryThreshold     uint64        `env:"PROTECTION_MEMORY_THRESHOLD" envDefault:"80"` // percentage
	MemoryCheckInterval time.Duration `env:"PROTECTION_MEMORY_CHECK_INTERVAL" envDefault:"1m"`

	// Token Bucket (Anti-flood)
	TokenBucketSize int64   `env:"PROTECTION_TOKEN_BUCKET_SIZE" envDefault:"100"`
	TokenFillRate   float64 `env:"PROTECTION_TOKEN_FILL_RATE" envDefault:"10"` // tokens per second
}

// Protection combines all protection mechanisms
type Protection struct {
	config   *ProtectionConfig
	slowReq  *SlowRequestProtector
	ipFilter *IPFilter
	failed   *FailedAttempts
	memory   *MemoryMonitor
	tokens   *TokenBucket
	mu       sync.RWMutex
}

// NewProtection creates a new instance of Protection with all mechanisms
func NewProtection(config *ProtectionConfig) *Protection {
	p := &Protection{
		config:   config,
		slowReq:  NewSlowRequestProtector(config.MinReadRate, config.ReadTimeout),
		ipFilter: NewIPFilter(config.IPWhitelist, config.IPBlacklist),
		failed:   NewFailedAttempts(config.MaxFailedAttempts, config.FailedBlockTime),
		memory:   NewMemoryMonitor(config.MemoryThreshold, config.MemoryCheckInterval),
		tokens:   NewTokenBucket(config.TokenBucketSize, config.TokenFillRate),
	}

	return p
}

// Start begins all protection mechanisms
func (p *Protection) Start() error {
	// Start memory monitoring
	go p.memory.Start()
	// Start token bucket refill
	go p.tokens.Start()
	return nil
}

// Stop gracefully stops all protection mechanisms
func (p *Protection) Stop() error {
	p.memory.Stop()
	p.tokens.Stop()
	return nil
}

// ProtectedConn wraps a connection with all protections
func (p *Protection) ProtectedConn(conn net.Conn) (net.Conn, error) {
	// Check IP restrictions
	if !p.ipFilter.IsAllowed(conn.RemoteAddr()) {
		return nil, ErrIPBlocked
	}

	// Check memory
	if p.memory.IsOverloaded() {
		return nil, ErrServerOverloaded
	}

	// Check failed attempts
	if p.failed.IsBlocked(conn.RemoteAddr()) {
		return nil, ErrTooManyFailures
	}

	// Apply slow request protection
	protected := p.slowReq.Protect(conn)

	return protected, nil
}

// RegisterFailure registers a failed attempt from an IP
func (p *Protection) RegisterFailure(addr net.Addr) {
	p.failed.RegisterFailure(addr)
}

// IsAllowed checks if a request should be allowed
func (p *Protection) IsAllowed(addr net.Addr) error {
	// Check IP restrictions
	if !p.ipFilter.IsAllowed(addr) {
		return ErrIPBlocked
	}

	// Check failed attempts
	if p.failed.IsBlocked(addr) {
		return ErrTooManyFailures
	}

	// Check memory
	if p.memory.IsOverloaded() {
		return ErrServerOverloaded
	}

	// Check token bucket (anti-flood)
	if !p.tokens.Take(addr.String()) {
		return ErrTooManyRequests
	}

	return nil
}

// Custom errors
var (
	ErrIPBlocked        = errors.New("ip is blocked")
	ErrServerOverloaded = errors.New("server is overloaded")
	ErrTooManyFailures  = errors.New("too many failed attempts")
	ErrTooManyRequests  = errors.New("too many requests")
	ErrSlowConnection   = errors.New("connection too slow")
)
