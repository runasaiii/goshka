package retry

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestIsRetryable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		resp *http.Response
		err  error
		want bool
	}{
		{
			name: "timeout error",
			err:  timeoutError{},
			want: true,
		},
		{
			name: "too many requests",
			resp: &http.Response{StatusCode: http.StatusTooManyRequests},
			err:  errors.New("retryable status"),
			want: true,
		},
		{
			name: "service unavailable",
			resp: &http.Response{StatusCode: http.StatusServiceUnavailable},
			err:  errors.New("retryable status"),
			want: true,
		},
		{
			name: "unauthorized",
			resp: &http.Response{StatusCode: http.StatusUnauthorized},
			err:  errors.New("non-retryable status"),
			want: false,
		},
		{
			name: "not found",
			resp: &http.Response{StatusCode: http.StatusNotFound},
			err:  errors.New("non-retryable status"),
			want: false,
		},
		{
			name: "bad request",
			resp: &http.Response{StatusCode: http.StatusBadRequest},
			err:  errors.New("non-retryable status"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsRetryable(tt.resp, tt.err)

			if got != tt.want {
				t.Fatalf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		attempt int
		max     time.Duration
	}{
		{name: "first attempt", attempt: 1, max: 500 * time.Millisecond},
		{name: "second attempt", attempt: 2, max: 1 * time.Second},
		{name: "capped attempt", attempt: 5, max: 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CalculateBackoff(tt.attempt)

			if got < 0 || got > tt.max {
				t.Fatalf("CalculateBackoff(%d) = %v, expected 0..%v", tt.attempt, got, tt.max)
			}
		})
	}
}

func TestPaymentClientExecutePaymentRetriesUntilSuccess(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attempts.Add(1)
		if attempt <= 3 {
			http.Error(w, "temporary outage", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := NewPaymentClient(server.URL)
	client.MaxRetries = 5
	client.BaseDelay = time.Millisecond
	client.MaxDelay = 5 * time.Millisecond
	client.backoffFn = func(int) time.Duration { return 0 }
	client.logf = func(string, ...any) {}

	result, err := client.ExecutePayment(context.Background())

	if err != nil {
		t.Fatalf("ExecutePayment() error = %v", err)
	}

	if result.Status != "success" {
		t.Fatalf("ExecutePayment() status = %q, want success", result.Status)
	}

	if got := attempts.Load(); got != 4 {
		t.Fatalf("ExecutePayment() attempts = %d, want 4", got)
	}
}

func TestPaymentClientExecutePaymentStopsOnNonRetryableStatus(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		http.Error(w, "invalid api key", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewPaymentClient(server.URL)
	client.MaxRetries = 5
	client.BaseDelay = time.Millisecond
	client.MaxDelay = 5 * time.Millisecond
	client.backoffFn = func(int) time.Duration { return 0 }
	client.logf = func(string, ...any) {}

	_, err := client.ExecutePayment(context.Background())
	if err == nil {
		t.Fatal("ExecutePayment() error = nil, want non-retryable failure")
	}

	if got := attempts.Load(); got != 1 {
		t.Fatalf("ExecutePayment() attempts = %d, want 1", got)
	}
}

func TestPaymentClientExecutePaymentStopsWhenContextExpires(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		http.Error(w, "temporary outage", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewPaymentClient(server.URL)
	client.MaxRetries = 5
	client.BaseDelay = 200 * time.Millisecond
	client.MaxDelay = 200 * time.Millisecond
	client.backoffFn = func(int) time.Duration { return 200 * time.Millisecond }
	client.logf = func(string, ...any) {}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := client.ExecutePayment(ctx)
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ExecutePayment() error = %v, want context deadline exceeded", err)
	}

	if got := attempts.Load(); got != 1 {
		t.Fatalf("ExecutePayment() attempts = %d, want 1", got)
	}

	if elapsed >= 180*time.Millisecond {
		t.Fatalf("ExecutePayment() waited too long after context cancellation: %v", elapsed)
	}
}
