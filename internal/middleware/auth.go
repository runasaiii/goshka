package middleware
import (
    "log"
    "net/http"
    "time"
)


func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r) 
        log.Printf("METHOD: %s | PATH: %s | DURATION: %v", r.Method, r.URL.Path, time.Since(start))
    })
}

func Auth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("X-API-KEY") != "roro123" {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte(`{"error": "Unauthorized"}`))
            return
        }
        next.ServeHTTP(w, r)
    })
}