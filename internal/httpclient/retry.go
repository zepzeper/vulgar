package httpclient

import (
	"bytes"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// retryMiddleware creates middleware that handles retries with exponential backoff
func (c *Client) retryMiddleware(next RoundTripperFunc) RoundTripperFunc {
	return func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		var err error
		var bodyBytes []byte

		// Save the body for retries
		if req.Body != nil {
			bodyBytes, err = io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			req.Body.Close()
		}

		for attempt := 0; attempt <= c.maxRetries; attempt++ {
			// Reset body for each attempt
			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}

			resp, err = next(req)

			// If successful or not retryable, return
			if err == nil && !c.shouldRetry(resp) {
				return resp, nil
			}

			// If this was the last attempt, return the error/response
			if attempt == c.maxRetries {
				if err != nil {
					return nil, err
				}
				return resp, nil
			}

			// Close response body before retry
			if resp != nil {
				resp.Body.Close()
			}

			// Calculate wait time with exponential backoff and jitter
			waitTime := c.calculateBackoff(attempt)
			time.Sleep(waitTime)
		}

		return resp, err
	}
}

// shouldRetry determines if a response should be retried
func (c *Client) shouldRetry(resp *http.Response) bool {
	if resp == nil {
		return true
	}

	for _, code := range c.retryOnStatus {
		if resp.StatusCode == code {
			return true
		}
	}

	return false
}

// calculateBackoff calculates the wait time for a retry attempt
func (c *Client) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: waitMin * 2^attempt
	wait := float64(c.retryWaitMin) * math.Pow(2, float64(attempt))

	// Add jitter (±25%)
	jitter := wait * 0.25 * (rand.Float64()*2 - 1)
	wait += jitter

	// Cap at maximum wait time
	if wait > float64(c.retryWaitMax) {
		wait = float64(c.retryWaitMax)
	}

	return time.Duration(wait)
}

// =============================================================================
// Rate Limiting
// =============================================================================

// rateLimiter implements a token bucket rate limiter
type rateLimiter struct {
	rate       float64   // tokens per second
	burst      int       // maximum burst size
	tokens     float64   // current token count
	lastUpdate time.Time // last time tokens were updated
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(requestsPerSecond float64) *rateLimiter {
	return &rateLimiter{
		rate:       requestsPerSecond,
		burst:      1,
		tokens:     1,
		lastUpdate: time.Now(),
	}
}

// newRateLimiterWithBurst creates a new rate limiter with burst capacity
func newRateLimiterWithBurst(requestsPerSecond float64, burst int) *rateLimiter {
	return &rateLimiter{
		rate:       requestsPerSecond,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Wait waits until a token is available
func (rl *rateLimiter) Wait() {
	for {
		if rl.tryAcquire() {
			return
		}
		// Wait a bit before trying again
		time.Sleep(time.Millisecond * 10)
	}
}

// tryAcquire tries to acquire a token, returns true if successful
func (rl *rateLimiter) tryAcquire() bool {
	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()

	// Add tokens based on elapsed time
	rl.tokens += elapsed * rl.rate
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}
	rl.lastUpdate = now

	// Try to consume a token
	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}

	return false
}

// rateLimitMiddleware creates middleware that applies rate limiting
func (c *Client) rateLimitMiddleware(next RoundTripperFunc) RoundTripperFunc {
	return func(req *http.Request) (*http.Response, error) {
		c.rateLimiter.Wait()
		return next(req)
	}
}
