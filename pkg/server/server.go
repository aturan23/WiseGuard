package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"wiseguard/pkg/logger"
	"wiseguard/pkg/pow"
	"wiseguard/pkg/quotes"
	"wiseguard/pkg/server/protection"
	"wiseguard/pkg/server/ratelimit"
)

// Config server configuration
type Config struct {
	Address           string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ShutdownTimeout   time.Duration
	MaxConnections    int
	InitialDifficulty uint8
}

// Server presents a PoW server
type Server struct {
	cfg          *Config
	log          logger.Logger
	listener     net.Listener
	powService   pow.Service
	quoteService quotes.Service
	rateLimiter  *ratelimit.IPRateLimiter
	conLimiter   *ratelimit.ConnectionLimiter
	protection   *protection.Protection

	// State
	activeConns  sync.WaitGroup
	currentConns atomic.Int32
	difficulty   atomic.Uint32
	shutdown     chan struct{}

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(cfg *Config, log logger.Logger, pow pow.Service, quotes quotes.Service, ctx context.Context) *Server {
	protectionCfg := &protection.ProtectionConfig{
		MinReadRate:         100,
		ReadTimeout:         5 * time.Second,
		EnableIPFilter:      true,
		MaxFailedAttempts:   5,
		FailedBlockTime:     15 * time.Minute,
		MemoryThreshold:     80,
		MemoryCheckInterval: time.Minute,
		TokenBucketSize:     100,
		TokenFillRate:       10,
	}

	srv := &Server{
		cfg:          cfg,
		log:          log.WithComponent("server"),
		powService:   pow,
		quoteService: quotes,
		ctx:          ctx,
		shutdown:     make(chan struct{}),
		rateLimiter:  ratelimit.NewIPRateLimiter(60, 10, 1*time.Hour),
		conLimiter:   ratelimit.NewConnectionLimiter(10, 1*time.Minute), // 10 connections per IP per minute
		protection:   protection.NewProtection(protectionCfg),
	}

	srv.difficulty.Store(uint32(cfg.InitialDifficulty))
	return srv
}

func (s *Server) Run() error {
	if err := s.protection.Start(); err != nil {
		return fmt.Errorf("failed to start protection: %w", err)
	}
	defer s.protection.Stop()

	var err error
	s.listener, err = net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}

	s.log.Info("server started", map[string]interface{}{
		"address": s.cfg.Address,
	})

	// Start PoW difficulty
	go s.adjustDifficulty()

	// Start accepting connections
	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return nil
				}
				s.log.Error("accept error", err, nil)
				continue
			}

			if err := s.protection.IsAllowed(conn.RemoteAddr()); err != nil {
				s.log.Info("connection rejected", map[string]interface{}{
					"remote_addr": conn.RemoteAddr().String(),
					"reason":      err.Error(),
				})
				conn.Close()
				continue
			}

			protected, err := s.protection.ProtectedConn(conn)
			if err != nil {
				s.log.Info("connection rejected", map[string]interface{}{
					"remote_addr": conn.RemoteAddr().String(),
					"reason":      err.Error(),
				})
				conn.Close()
				continue
			}

			// Check connection limit before accepting connection
			if !s.conLimiter.AllowConnection(conn.RemoteAddr()) {
				s.log.Info("connection rejected due to connection limit", map[string]interface{}{
					"remote_addr": conn.RemoteAddr().String(),
				})
				conn.Close()
				continue
			}

			// Check rate limit before accepting connection
			if !s.rateLimiter.AllowConnection(conn.RemoteAddr()) {
				s.log.Info("connection rejected due to rate limit", map[string]interface{}{
					"remote_addr": conn.RemoteAddr().String(),
				})
				conn.Close()
				continue
			}

			// Check if we reached max connections
			if s.currentConns.Load() >= int32(s.cfg.MaxConnections) {
				s.log.Info("max connections reached, dropping connection", map[string]interface{}{
					"remote_addr": conn.RemoteAddr().String(),
				})
				conn.Close()
				continue
			}

			s.activeConns.Add(1)
			s.currentConns.Add(1)

			go func() {
				defer func() {
					s.activeConns.Done()
					s.currentConns.Add(-1)
					s.conLimiter.RemoveConnection(conn.RemoteAddr())
					protected.Close()
					conn.Close()
				}()

				if err := s.handleConnection(conn); err != nil {
					s.log.Error("connection error", err, map[string]interface{}{
						"remote_addr": conn.RemoteAddr().String(),
					})
					if !errors.Is(err, net.ErrClosed) {
						s.protection.RegisterFailure(conn.RemoteAddr())
					}
				}
			}()
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) error {
	deadline := time.Now().Add(s.cfg.ReadTimeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}

	currentDifficulty := uint8(s.difficulty.Load())
	s.log.Debug("creating challenge", map[string]interface{}{
		"difficulty":  currentDifficulty,
		"remote_addr": conn.RemoteAddr().String(),
	})

	challenge, err := s.powService.CreateChallenge(currentDifficulty)
	if err != nil {
		s.log.Error("failed to create challenge", err, map[string]interface{}{
			"difficulty":  currentDifficulty,
			"remote_addr": conn.RemoteAddr().String(),
		})
		return s.sendError(conn, "INTERNAL_ERROR", "Failed to create challenge")
	}

	if err := s.sendChallenge(conn, challenge); err != nil {
		return err
	}

	solution, err := s.readSolution(conn)
	if err != nil {
		return err
	}

	if !s.powService.VerifySolution(challenge, solution) {
		return s.sendError(conn, "INVALID_SOLUTION", "Solution verification failed")
	}

	return s.sendQuote(conn)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("starting graceful shutdown", nil)

	// Сначала отменяем контекст
	if s.cancel != nil {
		s.cancel()
	}

	// Закрываем listener, если он существует
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.log.Error("failed to close listener", err, nil)
			// Продолжаем shutdown даже при ошибке
		}
	}

	// Создаем канал для сигнала о завершении
	done := make(chan struct{})

	go func() {
		s.activeConns.Wait()
		close(done)
	}()

	// Ждем завершения всех соединений или таймаута
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		s.log.Info("shutdown complete", nil)
		return nil
	}
}

// adjustDifficulty dynamic difficulty adjustment
func (s *Server) adjustDifficulty() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			currentConns := s.currentConns.Load()
			var newDifficulty uint32

			switch {
			case currentConns > int32(s.cfg.MaxConnections*8/10):
				newDifficulty = uint32(s.cfg.InitialDifficulty) + 2
			case currentConns > int32(s.cfg.MaxConnections*5/10):
				newDifficulty = uint32(s.cfg.InitialDifficulty) + 1
			default:
				newDifficulty = uint32(s.cfg.InitialDifficulty)
			}

			oldDifficulty := s.difficulty.Swap(newDifficulty)
			if oldDifficulty != newDifficulty {
				s.log.Info("difficulty adjusted", map[string]interface{}{
					"old_difficulty": oldDifficulty,
					"new_difficulty": newDifficulty,
					"connections":    currentConns,
				})
			}
		}
	}
}
