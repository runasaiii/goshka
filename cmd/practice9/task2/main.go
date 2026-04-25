package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"goshka/practice9/idempotency"
)

type paymentResult struct {
	Status        string `json:"status"`
	Amount        int    `json:"amount"`
	TransactionID string `json:"transaction_id"`
}

func main() {
	dbFile, err := os.CreateTemp("", "practice9-idempotency-*.db")
	if err != nil {
		fmt.Printf("create temp db: %v\n", err)
		return
	}
	defer os.Remove(dbFile.Name())
	_ = dbFile.Close()

	store, err := idempotency.NewDBStore(dbFile.Name())
	if err != nil {
		fmt.Printf("create database store: %v\n", err)
		return
	}
	defer store.Close()

	var executions atomic.Int32

	handler := idempotency.Middleware(store, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		run := executions.Add(1)
		fmt.Printf("Processing started (run #%d)\n", run)

		time.Sleep(2 * time.Second)

		result := paymentResult{
			Status:        "paid",
			Amount:        1000,
			TransactionID: fmt.Sprintf("uuid-%d", time.Now().UnixNano()),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Processing finished (run #%d): %s\n", run, result.TransactionID)
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()
	key := "demo-idempotency-key"

	type requestResult struct {
		index  int
		status int
		body   string
		err    error
	}

	results := make(chan requestResult, 8)
	var wg sync.WaitGroup

	for idx := 1; idx <= 8; idx++ {
		wg.Add(1)
		go func(requestIndex int) {
			defer wg.Done()

			resp, body, err := sendPaymentRequest(
				context.Background(),
				client,
				server.URL,
				key,
			)
			if err != nil {
				results <- requestResult{index: requestIndex, err: err}
				return
			}

			results <- requestResult{
				index:  requestIndex,
				status: resp.StatusCode,
				body:   body,
			}
		}(idx)
	}

	wg.Wait()
	close(results)

	for result := range results {
		if result.err != nil {
			fmt.Printf("request %d -> error: %v\n", result.index, result.err)
			continue
		}

		fmt.Printf("request %d -> %d %s\n", result.index, result.status, http.StatusText(result.status))
		if result.body != "" {
			fmt.Printf("request %d body -> %s\n", result.index, result.body)
		}
	}

	resp, body, err := sendPaymentRequest(context.Background(), client, server.URL, key)
	if err != nil {
		fmt.Printf("replay request error: %v\n", err)
		return
	}

	fmt.Printf("replay request -> %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Printf("replay body -> %s\n", body)
	fmt.Printf("business logic executed %d time(s)\n", executions.Load())
}

func sendPaymentRequest(
	ctx context.Context,
	client *http.Client,
	url string,
	key string,
) (*http.Response, string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url+"/pay",
		bytes.NewReader([]byte(`{"amount":1000}`)),
	)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var body bytes.Buffer
	if _, err := body.ReadFrom(resp.Body); err != nil {
		return nil, "", err
	}

	return resp, body.String(), nil
}
