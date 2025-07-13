package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Middleware func(http.Handler) http.Handler

func Add(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.Mutex
	r        = rate.Every(100 * time.Millisecond) // 10 req/sec
	burst    = 5
)

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(r, burst)
		limiters[ip] = limiter
	}
	return limiter
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			log.Printf("Rate limit exceeded for %s", ip)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
