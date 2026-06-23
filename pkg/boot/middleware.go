// middleware.go — Built-in middleware for KylixBoot.
//
// Middleware composes as: Logger → CORS → Auth → Handler
// Each middleware wraps the next handler in the chain.
package boot

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// Logger logs each request with method, path, status, and duration.
func Logger() Middleware {
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			start := time.Now()
			resp := next(req)
			status := 200
			if resp != nil {
				status = resp.Status
			}
			log.Printf("[%s] %s %s → %d (%s)",
				req.Request.Method,
				req.Request.URL.Path,
				req.Request.RemoteAddr,
				status,
				time.Since(start),
			)
			return resp
		}
	}
}

// Recover catches panics in handlers and returns a 500 response.
func Recover() Middleware {
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			var resp *Response
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("panic in handler: %v", r)
						resp = Text(500, "Internal Server Error")
					}
				}()
				resp = next(req)
			}()
			return resp
		}
	}
}

// CORS allows cross-origin requests from any origin.
// Pass specific origins to lock it down: CORS("https://kylix.top").
func CORS(allowedOrigins ...string) Middleware {
	allowed := "*"
	if len(allowedOrigins) > 0 {
		allowed = strings.Join(allowedOrigins, ", ")
	}
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			resp := next(req)
			if resp == nil {
				resp = Text(204, "")
			}
			resp.WithHeader("Access-Control-Allow-Origin", allowed)
			resp.WithHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			resp.WithHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
			return resp
		}
	}
}

// Auth checks the Authorization header against a validator function.
// validator(token) returns (userID, ok). On failure, returns 401.
func Auth(validator func(token string) (string, bool)) Middleware {
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			header := req.Header("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return JSON(401, map[string]string{"error": "missing or invalid Authorization header"})
			}
			token := strings.TrimPrefix(header, "Bearer ")
			if userID, ok := validator(token); ok {
				_ = userID // Could attach to req.Context() in future
				return next(req)
			}
			return JSON(401, map[string]string{"error": "invalid token"})
		}
	}
}

// RateLimit allows N requests per duration from each client IP.
func RateLimit(limit int, window time.Duration) Middleware {
	type bucket struct {
		count     int
		windowEnd time.Time
	}
	var (
		mu      sync.Mutex
		buckets = map[string]*bucket{}
	)
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			ip := req.Request.RemoteAddr
			mu.Lock()
			b, ok := buckets[ip]
			now := time.Now()
			if !ok || now.After(b.windowEnd) {
				b = &bucket{count: 0, windowEnd: now.Add(window)}
				buckets[ip] = b
			}
			b.count++
			currentCount := b.count
			mu.Unlock()

			if currentCount > limit {
				return JSON(429, map[string]string{"error": fmt.Sprintf("rate limit: %d/%v", limit, window)})
			}
			return next(req)
		}
	}
}

// RequestID adds a unique X-Request-ID header.
func RequestID() Middleware {
	var counter int64
	var mu sync.Mutex
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			mu.Lock()
			counter++
			id := fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), counter)
			mu.Unlock()
			resp := next(req)
			if resp != nil {
				resp.WithHeader("X-Request-ID", id)
			}
			return resp
		}
	}
}
