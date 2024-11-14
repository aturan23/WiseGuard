package protection

import (
	"runtime"
	"time"
)

type MemoryMonitor struct {
	threshold uint64
	interval  time.Duration
	stopCh    chan struct{}
}

func NewMemoryMonitor(threshold uint64, interval time.Duration) *MemoryMonitor {
	return &MemoryMonitor{
		threshold: threshold,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

func (m *MemoryMonitor) Start() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			// Check if memory usage exceeds threshold
			if float64(memStats.Alloc)/float64(memStats.Sys)*100 > float64(m.threshold) {
				// Trigger GC when memory usage is high
				runtime.GC()
			}
		case <-m.stopCh:
			return
		}
	}
}

func (m *MemoryMonitor) Stop() {
	close(m.stopCh)
}

func (m *MemoryMonitor) IsOverloaded() bool {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return float64(memStats.Alloc)/float64(memStats.Sys)*100 > float64(m.threshold)
}
