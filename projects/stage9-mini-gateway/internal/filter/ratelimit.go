package filter

import (
	"net/http"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/limiter"
)

// RateLimit 限流过滤器。
func RateLimit(bucket *limiter.TokenBucket) Filter {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) {
			if bucket != nil && !bucket.Allow() {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte("rate limited"))
				return
			}
			next(w, r)
		}
	}
}
