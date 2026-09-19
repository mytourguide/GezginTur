package middleware

import (
	"net/http"
	"sync"
	"time"
)

// rateBucket, IP basina basit kayan pencere (sliding window) sayaci.
type rateBucket struct {
	mu     sync.Mutex
	hits   []time.Time
	limit  int
	window time.Duration
}

var (
	bucketsMu sync.Mutex
	buckets   = map[string]*rateBucket{}
)

// RateLimit, IP basina dakikalik istek siniri uygular (brute-force korumasi).
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ip = xff // reverse proxy arkasında gercek istemci IP'si
			}
			bucketsMu.Lock()
			b, ok := buckets[ip]
			if !ok {
				b = &rateBucket{limit: limit, window: window}
				buckets[ip] = b
			}
			bucketsMu.Unlock()

			b.mu.Lock()
			now := time.Now()
			cutoff := now.Add(-window)
			kept := b.hits[:0]
			for _, t := range b.hits {
				if t.After(cutoff) {
					kept = append(kept, t)
				}
			}
			b.hits = kept
			if len(b.hits) >= b.limit {
				b.mu.Unlock()
				http.Error(w, `{"error":"cok fazla istek, lutfen bekleyin"}`, http.StatusTooManyRequests)
				return
			}
			b.hits = append(b.hits, now)
			b.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
