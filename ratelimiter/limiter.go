package ratelimiter

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Limiter struct {
	storage   *FileStorage
	buckets   map[string]*TokenBucket
	mu        sync.RWMutex
	capacity  int
	rate      int
	closeChan chan struct{}
}

func NewLimiter(storagePath string, capacity, rate int) *Limiter {
	storage := NewFileStorage(storagePath)
	buckets, _ := storage.Load()

	limiter := &Limiter{
		storage:   storage,
		buckets:   buckets,
		capacity:  capacity,
		rate:      rate,
		closeChan: make(chan struct{}),
	}

	go limiter.persistLoop(1 * time.Minute)
	return limiter
}

func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		if !l.Allow(clientIP) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, exists := l.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			Capacity:   l.capacity,
			Rate:       l.rate,
			Tokens:     l.capacity,
			LastRefill: time.Now(),
		}
		l.buckets[key] = bucket
	}

	return bucket.Take()
}

func (l *Limiter) persistLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mu.RLock()
			l.storage.Save(l.buckets)
			l.mu.RUnlock()
		case <-l.closeChan:
			return
		}
	}
}

func (l *Limiter) Close() {
	close(l.closeChan)
	l.storage.Save(l.buckets)
}
