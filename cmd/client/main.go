package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"wiseguard/pkg/client"
	"wiseguard/pkg/config"
	"wiseguard/pkg/logger"
	"wiseguard/pkg/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	log := logger.NewLogger(logger.Config{
		Level:  cfg.Logger.Level,
		Pretty: cfg.Logger.Pretty,
	})
	log = log.WithComponent("main")

	if err != nil {
		log.Fatal("failed to load config", err, nil)
	}

	cl := client.NewClient(&client.Config{
		ServerAddress:  cfg.Client.ServerAddress,
		ConnectTimeout: cfg.Client.ConnectTimeout,
		ReadTimeout:    cfg.Client.ReadTimeout,
		WriteTimeout:   cfg.Client.WriteTimeout,
		RetryConfig: utils.RetryConfig{
			MaxAttempts:  cfg.Client.MaxAttempts,
			InitialDelay: cfg.Client.RetryDelay,
			MaxDelay:     cfg.Client.MaxRetryDelay,
			Factor:       1.5,
		},
	}, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Signal processing
	go func() {
		sig := <-sigChan
		log.Info("received signal", map[string]interface{}{
			"signal": sig.String(),
		})
		cancel()
	}()

	// Connect to the server
	if err := cl.Connect(ctx); err != nil {
		log.Fatal("failed to connect", err, nil)
	}
	defer cl.Close()

	// Get a quote
	quote, err := cl.GetQuote(ctx)
	if err != nil {
		log.Fatal("failed to get quote", err, nil)
	}

	log.Info("received quote", map[string]interface{}{
		"text":   quote.Text,
		"author": quote.Author,
	})
}
