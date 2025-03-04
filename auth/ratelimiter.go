package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func NewRateLimiterMiddleware() func(next http.Handler) http.Handler {
	// Define the rate limit (10 requests per minute)
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  10,
	}

	// Use an in-memory store
	store := memory.NewStore()
	limiterInstance := limiter.New(store, rate)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request IP for limiting
			limiterCtx, err := limiterInstance.Get(r.Context(), r.RemoteAddr)
			if err != nil {
				http.Error(w, "Could not apply rate limiting", http.StatusInternalServerError)
				return
			}

			// Add rate limiting headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiterCtx.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limiterCtx.Remaining))

			if limiterCtx.Reached {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
