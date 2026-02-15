package filter

import "net/http"

// APIKeyAuth APIKey 认证过滤器。
func APIKeyAuth(expected string) Filter {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) {
			if expected == "" {
				next(w, r)
				return
			}
			if r.Header.Get("X-API-Key") != expected {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("unauthorized"))
				return
			}
			next(w, r)
		}
	}
}
