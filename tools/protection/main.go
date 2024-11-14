package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"wiseguard/pkg/client"
	"wiseguard/pkg/logger"
	"wiseguard/pkg/protocol"
)

type attackFunction func(ctx context.Context, clientID int, stats *stats, log logger.Logger)

var (
	attackType = flag.String("attack", "invalid_pow", "Type of attack (invalid_pow, connection_limit, failed_attempts, slowloris)")
	numClients = flag.Int("clients", 50, "Number of concurrent clients")
	duration   = flag.Duration("duration", 10*time.Second, "Test duration")
	serverAddr = flag.String("server", "localhost:8080", "Server address")
)

type stats struct {
	successRequests atomic.Int64
	failedRequests  atomic.Int64
	blockedRequests atomic.Int64
}

func simulateInvalidPoW(ctx context.Context, clientID int, stats *stats, log logger.Logger) {
	clientCfg := &client.Config{
		ServerAddress:  *serverAddr,
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			cl := client.NewClient(clientCfg, log)
			if err := cl.Connect(ctx); err != nil {
				stats.blockedRequests.Add(1)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			_ = &protocol.Message{
				Version: protocol.Version,
				Type:    protocol.TypeSolution,
				Payload: []byte(`{"prefix":"invalid","solution":"wrong","nonce":"invalid"}`),
			}

			_, err := cl.GetQuote(ctx)
			cl.Close()

			if err != nil {
				stats.failedRequests.Add(1)
			} else {
				stats.successRequests.Add(1)
			}

			time.Sleep(50 * time.Millisecond)
		}
	}
}

func simulateConnectionLimit(ctx context.Context, clientID int, stats *stats, log logger.Logger) {
	connections := make([]net.Conn, 0)
	defer func() {
		for _, conn := range connections {
			conn.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := net.DialTimeout("tcp", *serverAddr, 1*time.Second)
			if err != nil {
				stats.blockedRequests.Add(1)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			connections = append(connections, conn)
			stats.successRequests.Add(1)

			time.Sleep(500 * time.Millisecond)
		}
	}
}

func simulateFailedAttempts(ctx context.Context, clientID int, stats *stats, log logger.Logger) {
	clientCfg := &client.Config{
		ServerAddress:  *serverAddr,
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
	}

	attempts := 0
	maxAttempts := 5

	for {
		select {
		case <-ctx.Done():
			return
		default:
			cl := client.NewClient(clientCfg, log)
			err := cl.Connect(ctx)
			if err != nil {
				stats.blockedRequests.Add(1)
				attempts++
				if attempts >= maxAttempts {
					time.Sleep(5 * time.Second)
					attempts = 0
				}
				continue
			}

			_, err = cl.GetQuote(ctx)
			cl.Close()

			if err != nil {
				stats.failedRequests.Add(1)
				attempts++
			} else {
				stats.successRequests.Add(1)
				attempts = 0
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func simulateSlowloris(ctx context.Context, clientID int, stats *stats, log logger.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := net.DialTimeout("tcp", *serverAddr, 1*time.Second)
			if err != nil {
				stats.blockedRequests.Add(1)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			go func(c net.Conn) {
				defer c.Close()
				buffer := make([]byte, 1)

				for i := 0; i < 10; i++ {
					select {
					case <-ctx.Done():
						return
					default:
						_, err := c.Write(buffer)
						if err != nil {
							stats.failedRequests.Add(1)
							return
						}
						stats.successRequests.Add(1)
						time.Sleep(500 * time.Millisecond)
					}
				}
			}(conn)

			time.Sleep(200 * time.Millisecond)
		}
	}
}

func main() {
	flag.Parse()

	log := logger.NewLogger(logger.Config{
		Level:  "info",
		Pretty: true,
	})

	stats := &stats{}
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	var wg sync.WaitGroup
	fmt.Printf("Starting %s attack simulation with %d clients for %v\n", *attackType, *numClients, *duration)

	var attackFunc attackFunction
	switch *attackType {
	case "invalid_pow":
		attackFunc = simulateInvalidPoW
	case "connection_limit":
		attackFunc = simulateConnectionLimit
	case "failed_attempts":
		attackFunc = simulateFailedAttempts
	case "slowloris":
		attackFunc = simulateSlowloris
	default:
		fmt.Printf("Unknown attack type: %s\n", *attackType)
		return
	}

	for i := 0; i < *numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			attackFunc(ctx, clientID, stats, log)
		}(i)
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		lastSuccess := int64(0)
		lastFailed := int64(0)
		lastBlocked := int64(0)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				currentSuccess := stats.successRequests.Load()
				currentFailed := stats.failedRequests.Load()
				currentBlocked := stats.blockedRequests.Load()

				fmt.Printf("\rRPS - Success: %d/s | Failed: %d/s | Blocked: %d/s | Total - Success: %d | Failed: %d | Blocked: %d",
					currentSuccess-lastSuccess,
					currentFailed-lastFailed,
					currentBlocked-lastBlocked,
					currentSuccess,
					currentFailed,
					currentBlocked)

				lastSuccess = currentSuccess
				lastFailed = currentFailed
				lastBlocked = currentBlocked
			}
		}
	}()

	wg.Wait()

	fmt.Printf("\n\nFinal Statistics for %s attack:\n", *attackType)
	fmt.Printf("Total Successful Requests: %d\n", stats.successRequests.Load())
	fmt.Printf("Total Failed Requests: %d\n", stats.failedRequests.Load())
	fmt.Printf("Total Blocked Requests: %d\n", stats.blockedRequests.Load())

	successRate := float64(stats.successRequests.Load()) / float64(stats.successRequests.Load()+stats.failedRequests.Load()+stats.blockedRequests.Load()) * 100
	fmt.Printf("Success Rate: %.2f%%\n", successRate)
}
