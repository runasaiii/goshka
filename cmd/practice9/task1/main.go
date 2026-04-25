package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"goshka/practice9/retry"
)

func main() {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attempts.Add(1)
		if attempt <= 3 {
			fmt.Printf("Gateway attempt %d -> 503 Service Unavailable\n", attempt)
			http.Error(w, "temporary overload", http.StatusServiceUnavailable)
			return
		}

		fmt.Printf("Gateway attempt %d -> 200 OK\n", attempt)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := retry.NewPaymentClient(server.URL)
	client.MaxRetries = 5
	client.BaseDelay = 500 * time.Millisecond
	client.MaxDelay = 5 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.ExecutePayment(ctx)
	if err != nil {
		fmt.Printf("payment failed: %v\n", err)
		return
	}

	fmt.Printf("Final response: %+v\n", *result)
}
