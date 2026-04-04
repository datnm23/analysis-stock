package middleware

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var errService = errors.New("service unavailable")

func TestCircuitBreaker_StartsInClosedState(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed, got %v", cb.State())
	}
}

func TestCircuitBreaker_StaysClosedOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	for i := 0; i < 10; i++ {
		err := cb.Execute(func() error { return nil })
		if err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}

	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed after successes, got %v", cb.State())
	}
	if cb.Failures() != 0 {
		t.Errorf("expected 0 failures, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxFailures:      3,
		Timeout:          1 * time.Second,
		HalfOpenRequests: 1,
	}
	cb := NewCircuitBreaker(cfg)

	// Cause 3 failures
	for i := 0; i < 3; i++ {
		cb.Execute(func() error { return errService })
	}

	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen after 3 failures, got %v", cb.State())
	}

	// Next call should be rejected immediately
	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_ResetsFailuresOnSuccess(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxFailures:      3,
		Timeout:          1 * time.Second,
		HalfOpenRequests: 1,
	}
	cb := NewCircuitBreaker(cfg)

	// 2 failures, then a success
	cb.Execute(func() error { return errService })
	cb.Execute(func() error { return errService })
	cb.Execute(func() error { return nil }) // success resets

	if cb.Failures() != 0 {
		t.Errorf("expected 0 failures after success, got %d", cb.Failures())
	}
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed, got %v", cb.State())
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          50 * time.Millisecond,
		HalfOpenRequests: 1,
	}
	cb := NewCircuitBreaker(cfg)

	// Open the circuit
	cb.Execute(func() error { return errService })
	cb.Execute(func() error { return errService })

	if cb.State() != StateOpen {
		t.Fatalf("expected StateOpen, got %v", cb.State())
	}

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// State check should transition to half-open
	if cb.State() != StateHalfOpen {
		t.Errorf("expected StateHalfOpen after timeout, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenSuccessCloses(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          50 * time.Millisecond,
		HalfOpenRequests: 1,
	}
	cb := NewCircuitBreaker(cfg)

	// Open → wait → half-open
	cb.Execute(func() error { return errService })
	cb.Execute(func() error { return errService })
	time.Sleep(100 * time.Millisecond)

	// Success in half-open → close
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed after half-open success, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          50 * time.Millisecond,
		HalfOpenRequests: 1,
	}
	cb := NewCircuitBreaker(cfg)

	// Open → wait → half-open
	cb.Execute(func() error { return errService })
	cb.Execute(func() error { return errService })
	time.Sleep(100 * time.Millisecond)

	// Failure in half-open → back to open
	cb.Execute(func() error { return errService })

	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen after half-open failure, got %v", cb.State())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          10 * time.Second,
		HalfOpenRequests: 1,
	}
	cb := NewCircuitBreaker(cfg)

	// Open the circuit
	cb.Execute(func() error { return errService })
	cb.Execute(func() error { return errService })

	if cb.State() != StateOpen {
		t.Fatalf("expected StateOpen, got %v", cb.State())
	}

	// Manual reset
	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed after reset, got %v", cb.State())
	}
	if cb.Failures() != 0 {
		t.Errorf("expected 0 failures after reset, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_InFlightLimitsHalfOpen(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxFailures:      2,
		Timeout:          50 * time.Millisecond,
		HalfOpenRequests: 2,
	}
	cb := NewCircuitBreaker(cfg)

	// Open → wait → half-open
	cb.Execute(func() error { return errService })
	cb.Execute(func() error { return errService })
	time.Sleep(100 * time.Millisecond)

	// Track how many actually execute vs get rejected
	var executed atomic.Int32
	var rejected atomic.Int32
	var wg sync.WaitGroup

	// Fire 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cb.Execute(func() error {
				executed.Add(1)
				time.Sleep(20 * time.Millisecond) // Simulate work
				return nil
			})
			if errors.Is(err, ErrCircuitOpen) {
				rejected.Add(1)
			}
		}()
	}
	wg.Wait()

	exec := executed.Load()
	rej := rejected.Load()

	// Should allow at most HalfOpenRequests (2)
	if exec > int32(cfg.HalfOpenRequests) {
		t.Errorf("expected at most %d executions in half-open, got %d", cfg.HalfOpenRequests, exec)
	}
	if rej < 8 {
		t.Errorf("expected at least 8 rejections, got %d", rej)
	}

	t.Logf("Half-open: %d executed, %d rejected", exec, rej)
}

func TestCircuitBreakerState_String(t *testing.T) {
	tests := []struct {
		state CircuitBreakerState
		want  string
	}{
		{StateClosed, "CLOSED"},
		{StateOpen, "OPEN"},
		{StateHalfOpen, "HALF_OPEN"},
		{CircuitBreakerState(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := tt.state.String()
		if got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}
