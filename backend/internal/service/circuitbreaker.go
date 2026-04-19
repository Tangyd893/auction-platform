package service

import (
	"sync/atomic"
	"time"
)

// CircuitBreaker 熔断器
// 状态转换: Closed(正常) -> Open(熔断) -> HalfOpen(半开) -> Closed
type CircuitBreaker struct {
	failures  atomic.Int64
	threshold int64
	timeout   time.Duration
	state     atomic.Int32 // 0=Closed, 1=Open, 2=HalfOpen
}

const (
	StateClosed int32 = iota
	StateOpen
	StateHalfOpen
)

func NewCircuitBreaker(threshold int64, timeout time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
	}

	// 定时恢复到半开状态
	go func() {
		ticker := time.NewTicker(timeout)
		for range ticker.C {
			if cb.state.Load() == StateOpen {
				cb.state.Store(StateHalfOpen)
			}
		}
	}()

	return cb
}

func (cb *CircuitBreaker) Allow() bool {
	state := cb.state.Load()
	if state == StateClosed {
		return true
	}
	if state == StateOpen {
		return false
	}
	// HalfOpen: 只允许 1 个请求通过做探测
	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	if cb.state.Load() == StateHalfOpen {
		// 探测成功，关闭熔断器
		cb.state.Store(StateClosed)
		cb.failures.Store(0)
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	f := cb.failures.Add(1)
	if f >= cb.threshold {
		cb.state.Store(StateOpen)
	}
}

func (cb *CircuitBreaker) State() string {
	switch cb.state.Load() {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
