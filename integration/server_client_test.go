package integration

import (
	"context"
	"errors"
	"testing"
	"time"
	"wiseguard/pkg/client"
	"wiseguard/pkg/logger"
	"wiseguard/pkg/pow"
	"wiseguard/pkg/quotes"
	"wiseguard/pkg/server"
	"wiseguard/pkg/utils"
)

func TestServerClientInteraction(t *testing.T) {
	log := logger.NewLogger(logger.Config{
		Level:  "debug",
		Pretty: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverCfg := &server.Config{
		Address:           ":9090",
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		MaxConnections:    100,
		InitialDifficulty: 1, // Small complexity for tests
	}

	powService := pow.NewService(log, 5*time.Minute)
	quoteService := quotes.NewService(log)

	srv := server.NewServer(serverCfg, log, powService, quoteService, ctx)

	errChan := make(chan error, 1)

	// Run the server
	go func() {
		if err := srv.Run(); err != nil {
			errChan <- err
		}
	}()

	// Wait for the server to start
	time.Sleep(time.Second)

	clientCfg := &client.Config{
		ServerAddress:  "localhost:9090",
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		RetryConfig: utils.RetryConfig{
			MaxAttempts:  3,
			InitialDelay: time.Second,
			MaxDelay:     5 * time.Second,
			Factor:       1.5,
		},
	}

	cl := client.NewClient(clientCfg, log)

	if err := cl.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	quote, err := cl.GetQuote(ctx)
	if err != nil {
		t.Fatalf("Failed to get quote: %v", err)
	}

	if quote.Text == "" {
		t.Error("Quote text is empty")
	}
	if quote.Author == "" {
		t.Error("Quote author is empty")
	}

	if err := cl.Close(); err != nil {
		t.Errorf("Failed to close client: %v", err)
	}

	// Graceful shutdown
	cancel() // Cancel the main context

	// Create a new context for the server shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Wait for the server to stop
	select {
	case err := <-errChan:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("Server error: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		// Continue
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		t.Errorf("Failed to shutdown server: %v", err)
	}

	// Wait for the server to stop
	select {
	case <-shutdownCtx.Done():
		t.Error("Shutdown timeout exceeded")
	case err := <-errChan:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("Error during shutdown: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		// Successfully stopped
	}
}
