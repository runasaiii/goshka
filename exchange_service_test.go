package goshka

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExchangeService_GetRate(t *testing.T) {
	t.Run("successful scenario", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/convert", r.URL.Path)
			assert.Equal(t, "USD", r.URL.Query().Get("from"))
			assert.Equal(t, "EUR", r.URL.Query().Get("to"))

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"base":"USD","target":"EUR","rate":0.92}`)
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)

		rate, err := service.GetRate("USD", "EUR")

		require.NoError(t, err)
		assert.Equal(t, 0.92, rate)
	})

	t.Run("api business error", func(t *testing.T) {
		tests := []struct {
			name   string
			status int
		}{
			{name: "bad request", status: http.StatusBadRequest},
			{name: "not found", status: http.StatusNotFound},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.status)
					fmt.Fprint(w, `{"error":"invalid currency pair"}`)
				}))
				defer server.Close()

				service := NewExchangeService(server.URL)

				rate, err := service.GetRate("USD", "UNKNOWN")

				require.Error(t, err)
				assert.Equal(t, 0.0, rate)
				assert.EqualError(t, err, "api error: invalid currency pair")
			})
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"base":"USD","target":"EUR","rate":`)
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)

		rate, err := service.GetRate("USD", "EUR")

		require.Error(t, err)
		assert.Equal(t, 0.0, rate)
		assert.ErrorContains(t, err, "decode error")
	})

	t.Run("slow response timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"base":"USD","target":"EUR","rate":0.92}`)
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)
		service.Client.Timeout = 10 * time.Millisecond

		rate, err := service.GetRate("USD", "EUR")

		require.Error(t, err)
		assert.Equal(t, 0.0, rate)
		assert.ErrorContains(t, err, "network error")
	})

	t.Run("server panic internal server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recover() != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, `{"error":"internal server error"}`)
				}
			}()

			panic("server panic")
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)

		rate, err := service.GetRate("USD", "EUR")

		require.Error(t, err)
		assert.Equal(t, 0.0, rate)
		assert.EqualError(t, err, "api error: internal server error")
	})

	t.Run("empty body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)

		rate, err := service.GetRate("USD", "EUR")

		require.Error(t, err)
		assert.Equal(t, 0.0, rate)
		assert.ErrorContains(t, err, "decode error")
	})
}
