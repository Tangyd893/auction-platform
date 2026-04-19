package service

import (
	"sync"
	"sync/atomic"
	"time"
)

// ============ Rate Limiter (Per-User) ============

type RateLimiter struct {
	mu       sync.Mutex
	requests map[int64][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[int64][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(userID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	var valid []time.Time
	if existing, ok := rl.requests[userID]; ok {
		for _, t := range existing {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[userID] = valid
		return false
	}

	rl.requests[userID] = append(valid, now)
	return true
}

func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.window)
	for userID, times := range rl.requests {
		var valid []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, userID)
		} else {
			rl.requests[userID] = valid
		}
	}
}

// ============ Connection Manager ============

type ConnectionManager struct {
	activeStreams   int64 // 用 int64 而非 atomic.Int64，直接用 sync/atomic 函数
	maxStreams    int64
	bidsPerSecond *RateLimiter
	mu            sync.RWMutex
	nonceCache    map[string]time.Time
	nonceMu       sync.RWMutex
}

func NewConnectionManager(maxStreams int, bidsPerSecondLimit int) *ConnectionManager {
	cm := &ConnectionManager{
		maxStreams:    int64(maxStreams),
		bidsPerSecond: NewRateLimiter(bidsPerSecondLimit, 1*time.Second),
		nonceCache:    make(map[string]time.Time),
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			cm.cleanupNonceCache()
			cm.bidsPerSecond.Cleanup()
		}
	}()

	return cm
}

func (cm *ConnectionManager) AllowBid(userID int64) bool {
	return cm.bidsPerSecond.Allow(userID)
}

func (cm *ConnectionManager) AddStream() bool {
	for {
		current := atomic.LoadInt64(&cm.activeStreams)
		if current >= cm.maxStreams {
			return false
		}
		if atomic.CompareAndSwapInt64(&cm.activeStreams, current, current+1) {
			return true
		}
	}
}

func (cm *ConnectionManager) RemoveStream() {
	atomic.AddInt64(&cm.activeStreams, -1)
}

func (cm *ConnectionManager) ActiveStreams() int64 {
	return atomic.LoadInt64(&cm.activeStreams)
}

func (cm *ConnectionManager) ValidNonce(nonce string, ts int64) bool {
	if time.Now().Unix()-ts > 300 {
		return false
	}

	cm.nonceMu.Lock()
	defer cm.nonceMu.Unlock()

	if _, exists := cm.nonceCache[nonce]; exists {
		return false
	}

	cm.nonceCache[nonce] = time.Now()

	if len(cm.nonceCache) > 10000 {
		cm.cleanupNonceCacheLocked()
	}

	return true
}

func (cm *ConnectionManager) cleanupNonceCache() {
	cm.nonceMu.Lock()
	defer cm.nonceMu.Unlock()
	cm.cleanupNonceCacheLocked()
}

func (cm *ConnectionManager) cleanupNonceCacheLocked() {
	cutoff := time.Now().Add(-10 * time.Minute)
	for nonce, t := range cm.nonceCache {
		if t.Before(cutoff) {
			delete(cm.nonceCache, nonce)
		}
	}
}
