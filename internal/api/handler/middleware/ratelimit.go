package middleware

import (
"net/http"
"sync"
"time"
)

var (
rpsMu sync.Mutex
rps   = map[string]time.Time{}
)

func RateLimit(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
key := r.RemoteAddr
now := time.Now()
rpsMu.Lock()
last := rps[key]
if now.Sub(last) < 10*time.Millisecond {
rpsMu.Unlock()
http.Error(w, "too many requests", http.StatusTooManyRequests)
return
}
rps[key] = now
rpsMu.Unlock()
next.ServeHTTP(w, r)
})
}
