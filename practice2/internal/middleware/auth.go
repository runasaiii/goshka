package middleware
import (
	"log"
	"net/http"
	"time"
)


func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().Format(time.RFC3339)
		log.Printf("[%s] %s %s", now, r.Method, r.URL.Path)

		myPass := "rororo123"
		clientKey := r.Header.Get("X-API-KEY")

		if clientKey != myPass {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "You dont have an access! (wrong api key)"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}