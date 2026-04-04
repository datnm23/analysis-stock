package middleware

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed   CircuitBreakerState = iota // Normal operation
	StateOpen                               // Failing, reject requests
	StateHalfOpen                           // Testing if service recovered
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig configures the circuit breaker behavior
type CircuitBreakerConfig struct {
	MaxFailures      int           // Failures before opening circuit
	Timeout          time.Duration // How long to stay open before half-open
	HalfOpenRequests int           // Number of test requests in half-open state
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:      5,
		Timeout:          30 * time.Second,
		HalfOpenRequests: 2,
	}
}

// ErrCircuitOpen is returned when the circuit breaker is open
var ErrCircuitOpen = errors.New("circuit breaker is open: service unavailable")

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.Mutex
	state           CircuitBreakerState
	failures        int
	successes       int
	inFlight        int // Tracks in-flight requests in half-open state
	lastFailureTime time.Time
	config          CircuitBreakerConfig
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		state:  StateClosed,
		config: config,
	}
}

// Execute runs the given function through the circuit breaker.
// Thread-safe: uses inFlight counter to prevent race in half-open state.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Decrement in-flight for half-open tracking
	if cb.state == StateHalfOpen {
		cb.inFlight--
	}

	if err != nil {
		cb.recordFailureLocked()
	} else {
		cb.recordSuccessLocked()
	}

	return err
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Auto-transition from Open → HalfOpen if timeout elapsed
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.config.Timeout {
		cb.state = StateHalfOpen
		cb.successes = 0
		cb.inFlight = 0
	}

	return cb.state
}

// Failures returns the current failure count
func (cb *CircuitBreaker) Failures() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failures
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has elapsed → transition to half-open
		if time.Since(cb.lastFailureTime) > cb.config.Timeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			cb.inFlight = 0
			cb.inFlight++
			return true
		}
		return false

	case StateHalfOpen:
		// Allow limited requests — track in-flight to prevent race
		if cb.successes+cb.inFlight < cb.config.HalfOpenRequests {
			cb.inFlight++
			return true
		}
		return false
	}

	return false
}

// recordSuccessLocked must be called with cb.mu held
func (cb *CircuitBreaker) recordSuccessLocked() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0

	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.HalfOpenRequests {
			// Enough successes → close circuit
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.inFlight = 0
		}
	}
}

// recordFailureLocked must be called with cb.mu held
func (cb *CircuitBreaker) recordFailureLocked() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.state = StateOpen
		}

	case StateHalfOpen:
		// Any failure in half-open → back to open
		cb.state = StateOpen
		cb.successes = 0
		cb.inFlight = 0
	}
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.inFlight = 0
}
