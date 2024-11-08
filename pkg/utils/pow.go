package utils

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
)

// SolvePoW searches for a solution to the PoW problem
func SolvePoW(ctx context.Context, prefix string, difficulty uint8) (string, error) {
	numWorkers := runtime.NumCPU()
	resultCh := make(chan string, 1)
	errorCh := make(chan error, 1)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerNum int) {
			defer wg.Done()

			// Each worker starts from its own number and then increments by the number of workers
			for j := workerNum; ; j += numWorkers {
				select {
				case <-ctx.Done():
					return
				default:
					solution := fmt.Sprintf("%d", j)
					hash := CalculateHash(prefix + solution)

					if ValidateHashDifficulty(hash, difficulty) {
						select {
						case resultCh <- solution:
							return
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(resultCh)
		close(errorCh)
	}()

	select {
	case solution, ok := <-resultCh:
		if !ok {
			return "", errors.New("no solution found")
		}
		return solution, nil
	case err := <-errorCh:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
