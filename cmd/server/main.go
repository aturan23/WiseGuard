package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"wiseguard/pkg/config"
	"wiseguard/pkg/logger"
	"wiseguard/pkg/pow"
	"wiseguard/pkg/quotes"
	"wiseguard/pkg/server"
)

func main() {
	// Main context and channel for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	log := logger.NewLogger(logger.Config{
		Level:  cfg.Logger.Level,
		Pretty: cfg.Logger.Pretty,
	})
	log = log.WithComponent("main")

	// Creation of PoW and quote service
	powService := pow.NewService(log, cfg.PoW.ChallengeTTL)
	quoteService := quotes.NewService(log)

	srv := server.NewServer(&server.Config{
		Address:           cfg.Server.Address,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		ShutdownTimeout:   cfg.Server.ShutdownTimeout,
		MaxConnections:    cfg.Server.MaxConnections,
		InitialDifficulty: cfg.PoW.InitialDifficulty,
	}, log, powService, quoteService, ctx)

	// Start the server
	errChan := make(chan error, 1)
	go func() {
		log.Info("starting server", map[string]interface{}{
			"address": cfg.Server.Address,
		})
		errChan <- srv.Run()
	}()

	// Signal processing
	select {
	case err := <-errChan:
		if err != nil {
			log.Error("server error", err, nil)
			os.Exit(1)
		}
	case sig := <-sigChan:
		log.Info("received shutdown signal", map[string]interface{}{
			"signal": sig.String(),
		})
		cancel() // Stop the server
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", err, nil)
		os.Exit(1)
	}
	log.Info("server stopped gracefully", nil)
}
