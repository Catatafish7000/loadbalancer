package ratelimiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	Capacity   int        `json:"capacity"`
	Tokens     int        `json:"tokens"`
	Rate       int        `json:"rate"`
	LastRefill time.Time  `json:"last_refill"`
	mu         sync.Mutex `json:"-"`
}

func (tb *TokenBucket) Take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	if tb.Tokens > 0 {
		tb.Tokens--
		return true
	}
	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.LastRefill)
	tokensToAdd := int(elapsed.Seconds()) * tb.Rate

	if tokensToAdd > 0 {
		tb.Tokens += tokensToAdd
		if tb.Tokens > tb.Capacity {
			tb.Tokens = tb.Capacity
		}
		tb.LastRefill = now
	}
}
