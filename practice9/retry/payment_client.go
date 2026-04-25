package retry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"
)

const (
	defaultBaseDelay = 500 * time.Millisecond
	defaultMaxDelay  = 5 * time.Second
)

type PaymentResponse struct {
	Status string `json:"status"`
}

type PaymentClient struct {
	URL        string
	HTTPClient *http.Client
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration

	backoffFn func(attempt int) time.Duration
	logf      func(format string, args ...any)
}

func NewPaymentClient(url string) *PaymentClient {
	client := &PaymentClient{
		URL:        url,
		HTTPClient: &http.Client{Timeout: 3 * time.Second},
		MaxRetries: 5,
		BaseDelay:  defaultBaseDelay,
		MaxDelay:   defaultMaxDelay,
		logf: func(format string, args ...any) {
			fmt.Printf(format, args...)
		},
	}

	client.backoffFn = func(attempt int) time.Duration {
		return calculateBackoff(
			attempt,
			client.BaseDelay,
			client.MaxDelay,
			fullJitter,
		)
	}

	return client
}

func IsRetryable(resp *http.Response, err error) bool {
	if err != nil {
		var timeoutErr interface{ Timeout() bool }
		if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
			return true
		}
	}

	if resp == nil {
		return false
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	case http.StatusUnauthorized, http.StatusNotFound:
		return false
	default:
		return false
	}
}

func CalculateBackoff(attempt int) time.Duration {
	return calculateBackoff(
		attempt,
		defaultBaseDelay,
		defaultMaxDelay,
		fullJitter,
	)
}

func (c *PaymentClient) ExecutePayment(ctx context.Context) (*PaymentResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= c.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.URL,
			bytes.NewReader([]byte(`{"amount":1000}`)),
		)
		if err != nil {
			return nil, fmt.Errorf("build payment request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.HTTPClient.Do(req)
		if err == nil {
			result, decodeErr := decodePaymentResponse(resp)
			if decodeErr == nil {
				c.logf("Attempt %d: Success!\n", attempt)
				return result, nil
			}

			lastErr = decodeErr
		} else {
			lastErr = err
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if !IsRetryable(resp, err) {
			return nil, lastErr
		}

		if attempt == c.MaxRetries {
			break
		}

		backoff := c.backoffFn(attempt)
		c.logf("Attempt %d failed: waiting %v before retry...\n", attempt, backoff)

		if err := waitWithContext(ctx, backoff); err != nil {
			return nil, err
		}
	}

	if lastErr == nil {
		lastErr = errors.New("payment execution failed")
	}

	return nil, lastErr
}

func decodePaymentResponse(resp *http.Response) (*PaymentResponse, error) {
	if resp == nil {
		return nil, errors.New("payment gateway returned no response")
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("payment gateway returned status %d", resp.StatusCode)
	}

	var result PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode payment response: %w", err)
	}

	return &result, nil
}

func calculateBackoff(
	attempt int,
	baseDelay time.Duration,
	maxDelay time.Duration,
	jitterFn func(time.Duration) time.Duration,
) time.Duration {
	if attempt < 1 {
		attempt = 1
	}

	backoff := baseDelay
	for current := 1; current < attempt; current++ {
		if backoff >= maxDelay/2 {
			backoff = maxDelay
			break
		}
		backoff *= 2
	}

	if backoff > maxDelay {
		backoff = maxDelay
	}

	return jitterFn(backoff)
}

func fullJitter(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}

	return time.Duration(rand.Int64N(int64(max) + 1))
}

func waitWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return ctx.Err()
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
