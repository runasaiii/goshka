package idempotency

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	_ "modernc.org/sqlite"
)

const (
	statusProcessing = "processing"
	statusCompleted  = "completed"
)

type CachedResponse struct {
	Status     string
	StatusCode int
	Body       []byte
	Header     http.Header
}

type DBStore struct {
	db *sql.DB
}

func NewDBStore(name string) (*DBStore, error) {
	db, err := sql.Open("sqlite", name)
	if err != nil {
		return nil, fmt.Errorf("open database store: %w", err)
	}

	db.SetMaxOpenConns(1)

	store := &DBStore{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *DBStore) Close() error {
	return s.db.Close()
}

func (s *DBStore) Reserve(ctx context.Context, key string) (*CachedResponse, bool, error) {
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO idempotency_keys (idempotency_key, status) VALUES (?, ?) ON CONFLICT(idempotency_key) DO NOTHING`,
		key,
		statusProcessing,
	)
	if err != nil {
		return nil, false, fmt.Errorf("reserve idempotency key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("check reserve rows: %w", err)
	}

	if rowsAffected == 1 {
		return nil, true, nil
	}

	cached, err := s.Get(ctx, key)
	if err != nil {
		return nil, false, err
	}

	return cached, false, nil
}

func (s *DBStore) Get(ctx context.Context, key string) (*CachedResponse, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT status, response_code, response_body, response_headers
		FROM idempotency_keys
		WHERE idempotency_key = ?`,
		key,
	)

	var (
		status      string
		statusCode  sql.NullInt64
		body        []byte
		headerBytes []byte
	)

	if err := row.Scan(&status, &statusCode, &body, &headerBytes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get cached response: %w", err)
	}

	header := http.Header{}
	if len(headerBytes) > 0 {
		if err := json.Unmarshal(headerBytes, &header); err != nil {
			return nil, fmt.Errorf("decode cached headers: %w", err)
		}
	}

	return &CachedResponse{
		Status:     status,
		StatusCode: int(statusCode.Int64),
		Body:       append([]byte(nil), body...),
		Header:     header,
	}, nil
}

func (s *DBStore) Complete(ctx context.Context, key string, resp *CachedResponse) error {
	headerBytes, err := json.Marshal(resp.Header)
	if err != nil {
		return fmt.Errorf("encode cached headers: %w", err)
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE idempotency_keys
		SET status = ?, response_code = ?, response_body = ?, response_headers = ?, completed_at = CURRENT_TIMESTAMP
		WHERE idempotency_key = ?`,
		statusCompleted,
		resp.StatusCode,
		resp.Body,
		headerBytes,
		key,
	)
	if err != nil {
		return fmt.Errorf("complete idempotency key: %w", err)
	}

	return nil
}

func Middleware(store *DBStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
			return
		}

		cached, reserved, err := store.Reserve(r.Context(), key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !reserved {
			if cached != nil && cached.Status == statusCompleted {
				writeCachedResponse(w, cached)
				return
			}

			http.Error(w, "Duplicate request in progress", http.StatusConflict)
			return
		}

		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)

		response := &CachedResponse{
			Status:     statusCompleted,
			StatusCode: recorder.Code,
			Body:       append([]byte(nil), recorder.Body.Bytes()...),
			Header:     cloneHeader(recorder.Header()),
		}

		if err := store.Complete(context.Background(), key, response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeCachedResponse(w, response)
	})
}

func (s *DBStore) init() error {
	if _, err := s.db.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
		return fmt.Errorf("set busy timeout: %w", err)
	}

	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS idempotency_keys (
			idempotency_key TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			response_code INTEGER,
			response_body BLOB,
			response_headers BLOB,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create idempotency table: %w", err)
	}

	return nil
}

func writeCachedResponse(w http.ResponseWriter, cached *CachedResponse) {
	for key, values := range cached.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(cached.StatusCode)
	_, _ = w.Write(cached.Body)
}

func cloneHeader(header http.Header) http.Header {
	cloned := make(http.Header, len(header))
	for key, values := range header {
		cloned[key] = append([]string(nil), values...)
	}

	return cloned
}
