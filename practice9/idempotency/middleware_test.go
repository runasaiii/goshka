package idempotency

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *DBStore {
	t.Helper()

	store, err := NewDBStore(filepath.Join(t.TempDir(), "idempotency.db"))
	if err != nil {
		t.Fatalf("NewDBStore() error = %v", err)
	}

	t.Cleanup(func() {
		_ = store.Close()
	})

	return store
}

func TestMiddlewareRejectsMissingKey(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	handler := Middleware(store, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("business logic must not be called without Idempotency-Key")
	}))

	req := httptest.NewRequest(http.MethodPost, "/pay", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Middleware() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestMiddlewareReturnsConflictThenReplaysCompletedResponse(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32

	handler := Middleware(store, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		close(started)
		<-release
		_, _ = w.Write([]byte(`{"status":"paid","amount":1000,"transaction_id":"tx-1"}`))
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()
	key := "demo-key"

	errCh := make(chan error, 1)
	go func() {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/pay", bytes.NewReader([]byte(`{}`)))
		if err != nil {
			errCh <- err
			return
		}

		req.Header.Set("Idempotency-Key", key)

		resp, err := client.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errCh <- &statusError{got: resp.StatusCode, want: http.StatusOK}
			return
		}

		errCh <- nil
	}()

	<-started

	conflictReq := httptest.NewRequest(http.MethodPost, "/pay", bytes.NewReader([]byte(`{}`)))
	conflictReq.Header.Set("Idempotency-Key", key)
	conflictRec := httptest.NewRecorder()
	handler.ServeHTTP(conflictRec, conflictReq)

	if conflictRec.Code != http.StatusConflict {
		t.Fatalf("conflict status = %d, want %d", conflictRec.Code, http.StatusConflict)
	}

	close(release)

	if err := <-errCh; err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	replayReq := httptest.NewRequest(http.MethodPost, "/pay", bytes.NewReader([]byte(`{}`)))
	replayReq.Header.Set("Idempotency-Key", key)
	replayRec := httptest.NewRecorder()
	handler.ServeHTTP(replayRec, replayReq)

	if replayRec.Code != http.StatusOK {
		t.Fatalf("replay status = %d, want %d", replayRec.Code, http.StatusOK)
	}

	if replayRec.Body.String() != `{"status":"paid","amount":1000,"transaction_id":"tx-1"}` {
		t.Fatalf("replay body = %q", replayRec.Body.String())
	}

	if replayRec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("replay content-type = %q, want application/json", replayRec.Header().Get("Content-Type"))
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("business logic calls = %d, want 1", got)
	}
}

func TestMiddlewareConcurrentRequestsExecuteBusinessLogicOnce(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	var calls atomic.Int32
	handler := Middleware(store, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{"status":"paid","amount":1000,"transaction_id":"tx-concurrent"}`))
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()
	key := "same-key"

	type result struct {
		status int
		body   map[string]any
		err    error
	}

	results := make(chan result, 8)
	var wg sync.WaitGroup

	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/pay", bytes.NewReader([]byte(`{}`)))
			if err != nil {
				results <- result{err: err}
				return
			}

			req.Header.Set("Idempotency-Key", key)

			resp, err := client.Do(req)
			if err != nil {
				results <- result{err: err}
				return
			}
			defer resp.Body.Close()

			var body map[string]any
			if resp.StatusCode == http.StatusOK {
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					results <- result{err: err}
					return
				}
			}

			results <- result{status: resp.StatusCode, body: body}
		}()
	}

	wg.Wait()
	close(results)

	successes := 0
	conflicts := 0

	for res := range results {
		if res.err != nil {
			t.Fatalf("concurrent request error = %v", res.err)
		}

		switch res.status {
		case http.StatusOK:
			successes++
		case http.StatusConflict:
			conflicts++
		default:
			t.Fatalf("unexpected status = %d", res.status)
		}
	}

	if successes != 1 {
		t.Fatalf("success count = %d, want 1", successes)
	}

	if conflicts != 7 {
		t.Fatalf("conflict count = %d, want 7", conflicts)
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("business logic calls = %d, want 1", got)
	}
}

type statusError struct {
	got  int
	want int
}

func (e *statusError) Error() string {
	return http.StatusText(e.got)
}
