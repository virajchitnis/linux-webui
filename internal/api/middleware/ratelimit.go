package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	tokens   float64
	lastFill time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64 // tokens per second
	capacity float64
}

func NewRateLimiter(requestsPerSecond, capacity float64) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     requestsPerSecond,
		capacity: capacity,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b, ok := rl.buckets[ip]
	if !ok {
		b = &bucket{tokens: rl.capacity, lastFill: time.Now()}
		rl.buckets[ip] = b
	}
	now := time.Now()
	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.lastFill = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !rl.Allow(ip) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, b := range rl.buckets {
			if b.lastFill.Before(cutoff) {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// WSLimiter tracks concurrent WebSocket connections per IP.
type WSLimiter struct {
	mu    sync.Mutex
	conns map[string]int
	max   int
}

func NewWSLimiter(maxPerIP int) *WSLimiter {
	return &WSLimiter{conns: make(map[string]int), max: maxPerIP}
}

func (wl *WSLimiter) Acquire(ip string) bool {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	if wl.conns[ip] >= wl.max {
		return false
	}
	wl.conns[ip]++
	return true
}

func (wl *WSLimiter) Release(ip string) {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	if wl.conns[ip] > 0 {
		wl.conns[ip]--
	}
}

func ClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
