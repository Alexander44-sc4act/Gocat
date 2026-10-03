// Package backoff provides reconnection backoff strategies
package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Strategy represents a backoff strategy
type Strategy string

const (
	// StrategyLinear increases delay linearly
	StrategyLinear Strategy = "linear"

	// StrategyExponential increases delay exponentially
	StrategyExponential Strategy = "exponential"

	// StrategyFibonacci uses fibonacci sequence for delays
	StrategyFibonacci Strategy = "fibonacci"

	// StrategyConstant uses constant delay
	StrategyConstant Strategy = "constant"
)

// Config holds backoff configuration
type Config struct {
	Strategy    Strategy
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	Jitter      bool
	MaxAttempts int // 0 for infinite
}

// DefaultConfig returns default backoff configuration
func DefaultConfig() *Config {
	return &Config{
		Strategy:    StrategyExponential,
		BaseDelay:   1 * time.Second,
		MaxDelay:    60 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
		MaxAttempts: 0, // Infinite by default
	}
}

// Calculator calculates backoff delays
type Calculator struct {
	config *Config
	fib1   int
	fib2   int
}

// New creates a new backoff calculator
func New(config *Config) *Calculator {
	if config == nil {
		config = DefaultConfig()
	}
	return &Calculator{
		config: config,
		fib1:   1,
		fib2:   1,
	}
}

// Next returns the next backoff duration for the given attempt
func (c *Calculator) Next(attempt int) time.Duration {
	var delay time.Duration

	switch c.config.Strategy {
	case StrategyLinear:
		delay = c.linear(attempt)
	case StrategyExponential:
		delay = c.exponential(attempt)
	case StrategyFibonacci:
		delay = c.fibonacci(attempt)
	case StrategyConstant:
		delay = c.config.BaseDelay
	default:
		delay = c.exponential(attempt)
	}

	// Apply max delay cap
	if delay > c.config.MaxDelay {
		delay = c.config.MaxDelay
	}

	// Apply jitter if enabled
	if c.config.Jitter {
		delay = c.addJitter(delay)
	}

	return delay
}

// ShouldRetry returns whether to retry based on attempt count
func (c *Calculator) ShouldRetry(attempt int) bool {
	if c.config.MaxAttempts == 0 {
		return true // Infinite retries
	}
	return attempt < c.config.MaxAttempts
}

// linear implements linear backoff
func (c *Calculator) linear(attempt int) time.Duration {
	return c.config.BaseDelay * time.Duration(attempt+1)
}

// exponential implements exponential backoff
func (c *Calculator) exponential(attempt int) time.Duration {
	multiplier := math.Pow(c.config.Multiplier, float64(attempt))
	return time.Duration(float64(c.config.BaseDelay) * multiplier)
}

// fibonacci implements fibonacci backoff
func (c *Calculator) fibonacci(attempt int) time.Duration {
	if attempt == 0 {
		c.fib1 = 1
		c.fib2 = 1
		return c.config.BaseDelay
	}

	next := c.fib1 + c.fib2
	c.fib1 = c.fib2
	c.fib2 = next

	return c.config.BaseDelay * time.Duration(next)
}

// addJitter adds random jitter to prevent thundering herd
func (c *Calculator) addJitter(delay time.Duration) time.Duration {
	// Add ±25% jitter
	jitterRange := float64(delay) * 0.25
	jitter := (rand.Float64() * 2 * jitterRange) - jitterRange

	newDelay := time.Duration(float64(delay) + jitter)

	// Ensure delay is not negative
	if newDelay < 0 {
		return delay
	}

	return newDelay
}

// Reset resets the backoff state (useful for fibonacci)
func (c *Calculator) Reset() {
	c.fib1 = 1
	c.fib2 = 1
}
