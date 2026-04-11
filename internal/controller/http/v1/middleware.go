package v1
import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"practice-7/utils"

	"github.com/gin-gonic/gin"
)


const (
	CtxJWTUserID = "jwt_user_id"
	CtxJWTRole   = "jwt_role"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, role, err := utils.ParseBearerAccessToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(CtxJWTUserID, userID)
		c.Set(CtxJWTRole, role)
		c.Next()
	}
}

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(CtxJWTRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		role, ok := v.(string)
		if !ok || role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

type rateBucket struct {
	count int
	until time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateBucket
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{buckets: make(map[string]*rateBucket)}
}

func (rl *rateLimiter) allow(key string, max int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok || now.After(b.until) {
		rl.buckets[key] = &rateBucket{count: 1, until: now.Add(window)}
		return true
	}
	if b.count >= max {
		return false
	}
	b.count++
	return true
}

func normalizeIP(ip string) string {
	switch strings.TrimSpace(strings.ToLower(ip)) {
	case "::1", "127.0.0.1", "0:0:0:0:0:0:0:1", "::ffff:127.0.0.1":
		return "loopback"
	case "":
		return "unknown"
	default:
		return ip
	}
}

func rateLimitKey(c *gin.Context) string {
	if uid, err := utils.ParseBearerAnyUserID(c.GetHeader("Authorization")); err == nil && uid != "" {
		return "user:" + uid
	}
	return "ip:" + normalizeIP(c.ClientIP())
}

func RateLimitMiddleware(max int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter()
	return func(c *gin.Context) {
		if !rl.allow(rateLimitKey(c), max, window) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}

var (
	defaultRateLimitOnce sync.Once
	defaultRateLimitMW   gin.HandlerFunc
)

func DefaultRateLimit() gin.HandlerFunc {
	defaultRateLimitOnce.Do(func() {
		max := 20
		if s := os.Getenv("RATE_LIMIT_MAX"); s != "" {
			if n, err := strconv.Atoi(s); err == nil && n > 0 {
				max = n
			}
		}
		sec := 60
		if s := os.Getenv("RATE_LIMIT_WINDOW_SEC"); s != "" {
			if n, err := strconv.Atoi(s); err == nil && n > 0 {
				sec = n
			}
		}
		win := time.Duration(sec) * time.Second
		log.Printf("rate limit: max %d requests per %v (set RATE_LIMIT_MAX / RATE_LIMIT_WINDOW_SEC to override)", max, win)
		defaultRateLimitMW = RateLimitMiddleware(max, win)
	})
	return defaultRateLimitMW
}
