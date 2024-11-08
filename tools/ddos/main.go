package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"wiseguard/pkg/client"
	"wiseguard/pkg/logger"
	"wiseguard/pkg/utils"
)

var (
	numClients     = flag.Int("clients", 100, "Number of concurrent clients")
	duration       = flag.Duration("duration", 30*time.Second, "Test duration")
	reportInterval = flag.Duration("report", 1*time.Second, "Report interval")
	serverAddr     = flag.String("server", "localhost:8080", "Server address")
)

type stats struct {
	successRequests atomic.Int64
	failedRequests  atomic.Int64
	totalLatency    atomic.Int64
}

func main() {
	flag.Parse()

	log := logger.NewLogger(logger.Config{
		Level:  "info",
		Pretty: true,
	})

	// Create stats object
	stats := &stats{}

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	// Create a goroutine to report stats
	go reportStats(ctx, stats)

	var wg sync.WaitGroup
	fmt.Printf("Starting DDoS simulation with %d clients for %v\n", *numClients, *duration)

	for i := 0; i < *numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			runClient(ctx, clientID, stats, log)
		}(i)
	}

	// Wait for all clients to finish
	wg.Wait()

	// Print final stats
	printFinalStats(stats, *duration)
}

func runClient(ctx context.Context, id int, stats *stats, log logger.Logger) {
	clientCfg := &client.Config{
		ServerAddress:  *serverAddr,
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		RetryConfig: utils.RetryConfig{
			MaxAttempts:  1, // Disable retries
			InitialDelay: time.Second,
			MaxDelay:     5 * time.Second,
			Factor:       1.5,
		},
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			startTime := time.Now()

			// Create new client for each request
			cl := client.NewClient(clientCfg, log)

			// Try to connect and get a quote
			err := cl.Connect(ctx)
			if err == nil {
				_, err = cl.GetQuote(ctx)
			}
			cl.Close()

			// Update stats
			if err != nil {
				stats.failedRequests.Add(1)
				log.Debug("request failed", map[string]interface{}{
					"client_id": id,
					"error":     err.Error(),
				})
			} else {
				stats.successRequests.Add(1)
				latency := time.Since(startTime).Milliseconds()
				stats.totalLatency.Add(latency)
			}

			// Small delay before sending next request
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func reportStats(ctx context.Context, stats *stats) {
	ticker := time.NewTicker(*reportInterval)
	defer ticker.Stop()

	var lastSuccess, lastFailed int64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentSuccess := stats.successRequests.Load()
			currentFailed := stats.failedRequests.Load()

			successDelta := currentSuccess - lastSuccess
			failedDelta := currentFailed - lastFailed

			rps := float64(successDelta+failedDelta) / reportInterval.Seconds()

			avgLatency := float64(0)
			if currentSuccess > 0 {
				avgLatency = float64(stats.totalLatency.Load()) / float64(currentSuccess)
			}

			fmt.Printf("\rRPS: %.2f | Success: %d (%d/s) | Failed: %d (%d/s) | Avg Latency: %.2fms",
				rps,
				currentSuccess,
				successDelta,
				currentFailed,
				failedDelta,
				avgLatency,
			)

			lastSuccess = currentSuccess
			lastFailed = currentFailed
		}
	}
}

func printFinalStats(stats *stats, duration time.Duration) {
	totalRequests := stats.successRequests.Load() + stats.failedRequests.Load()
	successRate := float64(stats.successRequests.Load()) / float64(totalRequests) * 100
	avgLatency := float64(0)
	if stats.successRequests.Load() > 0 {
		avgLatency = float64(stats.totalLatency.Load()) / float64(stats.successRequests.Load())
	}
	rps := float64(totalRequests) / duration.Seconds()

	fmt.Printf("\n\nFinal Statistics:\n")
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Successful Requests: %d\n", stats.successRequests.Load())
	fmt.Printf("Failed Requests: %d\n", stats.failedRequests.Load())
	fmt.Printf("Success Rate: %.2f%%\n", successRate)
	fmt.Printf("Average Latency: %.2fms\n", avgLatency)
	fmt.Printf("Average RPS: %.2f\n", rps)
}
