package protection

import (
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	capacity := int64(10)
	fillRate := float64(1) // 1 token per second

	tb := NewTokenBucket(capacity, fillRate)
	tb.Start()
	defer tb.Stop()

	ip := "127.0.0.1"

	// Test 1: Initial capacity
	if tokens := tb.GetTokens(ip); tokens != capacity {
		t.Errorf("Expected %d tokens, got %d", capacity, tokens)
	}

	// Test 2: Take tokens
	for i := 0; i < int(capacity); i++ {
		if !tb.Take(ip) {
			t.Errorf("Failed to take token %d", i)
		}
	}

	// Test 3: Should be empty
	if tb.Take(ip) {
		t.Error("Should not be able to take token when empty")
	}

	// Test 4: Refill
	time.Sleep(time.Second * 2)
	// Should have 2 new tokens after 2 seconds
	if tokens := tb.GetTokens(ip); tokens != 2 {
		t.Errorf("Expected 2 tokens after refill, got %d", tokens)
	}

	// Test 5: Cleanup
	time.Sleep(tb.cleanupInterval + time.Second)
	// After cleanup interval, bucket should be removed and reset to full capacity
	if tokens := tb.GetTokens(ip); tokens != capacity {
		t.Errorf("Expected %d tokens after cleanup, got %d", capacity, tokens)
	}
}
