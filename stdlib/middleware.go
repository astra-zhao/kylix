package stdlib

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// CORSMiddleware adds Cross-Origin Resource Sharing headers
type CORSMiddleware struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// NewCORSMiddleware creates a new CORS middleware with default settings
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{},
		MaxAge:         86400,
	}
}

// SetAllowedOrigins sets the allowed origins
func (m *CORSMiddleware) SetAllowedOrigins(origins []string) *CORSMiddleware {
	m.AllowedOrigins = origins
	return m
}

// SetAllowedMethods sets the allowed HTTP methods
func (m *CORSMiddleware) SetAllowedMethods(methods []string) *CORSMiddleware {
	m.AllowedMethods = methods
	return m
}

// SetAllowedHeaders sets the allowed headers
func (m *CORSMiddleware) SetAllowedHeaders(headers []string) *CORSMiddleware {
	m.AllowedHeaders = headers
	return m
}

// SetAllowCredentials sets whether credentials are allowed
func (m *CORSMiddleware) SetAllowCredentials(allow bool) *CORSMiddleware {
	m.AllowCredentials = allow
	return m
}

// Handle processes the CORS middleware
func (m *CORSMiddleware) Handle(req *TRequest, res *TResponse) bool {
	origin := req.Header("Origin")

	// Check if origin is allowed
	allowed := false
	for _, o := range m.AllowedOrigins {
		if o == "*" || o == origin {
			allowed = true
			break
		}
	}

	if !allowed {
		return true // Continue to next middleware
	}

	// Set CORS headers
	if len(m.AllowedOrigins) == 1 && m.AllowedOrigins[0] == "*" {
		res.Header("Access-Control-Allow-Origin", "*")
	} else {
		res.Header("Access-Control-Allow-Origin", origin)
		res.Header("Vary", "Origin")
	}

	res.Header("Access-Control-Allow-Methods", strings.Join(m.AllowedMethods, ", "))
	res.Header("Access-Control-Allow-Headers", strings.Join(m.AllowedHeaders, ", "))

	if len(m.ExposedHeaders) > 0 {
		res.Header("Access-Control-Expose-Headers", strings.Join(m.ExposedHeaders, ", "))
	}

	if m.AllowCredentials {
		res.Header("Access-Control-Allow-Credentials", "true")
	}

	if m.MaxAge > 0 {
		res.Header("Access-Control-Max-Age", fmt.Sprintf("%d", m.MaxAge))
	}

	// Handle preflight request
	if req.Method() == "OPTIONS" {
		res.Status(204)
		res.Send("")
		return false // Stop processing
	}

	return true // Continue to next middleware
}

// AuthMiddleware provides basic authentication
type AuthMiddleware struct {
	ValidateToken func(token string) bool
	HeaderName    string
	TokenPrefix   string
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(validateToken func(string) bool) *AuthMiddleware {
	return &AuthMiddleware{
		ValidateToken: validateToken,
		HeaderName:    "Authorization",
		TokenPrefix:   "Bearer ",
	}
}

// SetHeaderName sets the header name to check for the token
func (m *AuthMiddleware) SetHeaderName(name string) *AuthMiddleware {
	m.HeaderName = name
	return m
}

// SetTokenPrefix sets the token prefix (e.g., "Bearer ")
func (m *AuthMiddleware) SetTokenPrefix(prefix string) *AuthMiddleware {
	m.TokenPrefix = prefix
	return m
}

// Handle processes the authentication middleware
func (m *AuthMiddleware) Handle(req *TRequest, res *TResponse) bool {
	authHeader := req.Header(m.HeaderName)

	if authHeader == "" {
		res.Status(401)
		res.Header("WWW-Authenticate", "Bearer")
		res.JSON(map[string]interface{}{
			"error":   "unauthorized",
			"message": "Missing authentication token",
		})
		return false // Stop processing
	}

	token := authHeader
	if m.TokenPrefix != "" {
		if !strings.HasPrefix(authHeader, m.TokenPrefix) {
			res.Status(401)
			res.JSON(map[string]interface{}{
				"error":   "unauthorized",
				"message": "Invalid token format",
			})
			return false
		}
		token = strings.TrimPrefix(authHeader, m.TokenPrefix)
	}

	if m.ValidateToken != nil && !m.ValidateToken(token) {
		res.Status(401)
		res.JSON(map[string]interface{}{
			"error":   "unauthorized",
			"message": "Invalid token",
		})
		return false
	}

	// Store token in request context (as a param for simplicity)
	req.Params["__auth_token"] = token

	return true // Continue to next middleware
}

// RequestIDMiddleware adds a unique request ID to each request
type RequestIDMiddleware struct {
	HeaderName string
}

// NewRequestIDMiddleware creates a new request ID middleware
func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{
		HeaderName: "X-Request-ID",
	}
}

// SetHeaderName sets the header name for the request ID
func (m *RequestIDMiddleware) SetHeaderName(name string) *RequestIDMiddleware {
	m.HeaderName = name
	return m
}

// generateRequestID generates a random request ID
func generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Handle processes the request ID middleware
func (m *RequestIDMiddleware) Handle(req *TRequest, res *TResponse) bool {
	// Check if request already has an ID
	requestID := req.Header(m.HeaderName)

	if requestID == "" {
		// Generate a new request ID
		requestID = generateRequestID()
	}

	// Store in request params
	req.Params["__request_id"] = requestID

	// Add to response headers
	res.Header(m.HeaderName, requestID)

	return true // Continue to next middleware
}

// RateLimitMiddleware provides rate limiting
type RateLimitMiddleware struct {
	MaxRequests int
	Window      time.Duration
	clients     map[string]*clientRate
	mu          sync.Mutex
}

type clientRate struct {
	count    int
	lastSeen time.Time
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(maxRequests int, window time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		MaxRequests: maxRequests,
		Window:      window,
		clients:     make(map[string]*clientRate),
	}
}

// getClientIP extracts the client IP from the request
func (m *RateLimitMiddleware) getClientIP(req *TRequest) string {
	// Check X-Forwarded-For header first
	forwarded := req.Header("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	realIP := req.Header("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr (without port)
	addr := req.Request.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// Handle processes the rate limiting middleware
func (m *RateLimitMiddleware) Handle(req *TRequest, res *TResponse) bool {
	clientIP := m.getClientIP(req)

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	client, exists := m.clients[clientIP]

	if !exists || now.Sub(client.lastSeen) > m.Window {
		// New client or window expired
		m.clients[clientIP] = &clientRate{
			count:    1,
			lastSeen: now,
		}
		return true // Continue to next middleware
	}

	// Increment counter
	client.count++
	client.lastSeen = now

	// Check if rate limit exceeded
	if client.count > m.MaxRequests {
		// Calculate retry after
		retryAfter := m.Window - now.Sub(client.lastSeen)

		res.Status(429)
		res.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
		res.Header("X-RateLimit-Limit", fmt.Sprintf("%d", m.MaxRequests))
		res.Header("X-RateLimit-Remaining", "0")
		res.JSON(map[string]interface{}{
			"error":   "rate_limit_exceeded",
			"message": "Too many requests",
		})
		return false // Stop processing
	}

	// Add rate limit headers
	remaining := m.MaxRequests - client.count
	res.Header("X-RateLimit-Limit", fmt.Sprintf("%d", m.MaxRequests))
	res.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

	return true // Continue to next middleware
}

// LoggingMiddleware logs request details
type LoggingMiddleware struct {
	Format string // "simple", "combined", "custom"
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		Format: "simple",
	}
}

// SetFormat sets the log format
func (m *LoggingMiddleware) SetFormat(format string) *LoggingMiddleware {
	m.Format = format
	return m
}

// Handle processes the logging middleware
func (m *LoggingMiddleware) Handle(req *TRequest, res *TResponse) bool {
	// Get request ID if available
	requestID := req.Params["__request_id"]
	if requestID == "" {
		requestID = "-"
	}

	// Log after response is sent (we'll log before for simplicity)
	fmt.Printf("[%s] %s %s %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		requestID,
		req.Method(),
		req.Path())

	return true // Continue to next middleware
}

// RecoveryMiddleware recovers from panics
type RecoveryMiddleware struct {
	LogErrors bool
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{
		LogErrors: true,
	}
}

// SetLogErrors sets whether to log errors
func (m *RecoveryMiddleware) SetLogErrors(log bool) *RecoveryMiddleware {
	m.LogErrors = log
	return m
}

// Handle processes the recovery middleware
func (m *RecoveryMiddleware) Handle(req *TRequest, res *TResponse) bool {
	// Note: This is a simplified version. In real implementation,
	// you'd use defer/recover in the actual handler execution
	return true // Continue to next middleware
}

// Helper function to get request ID from request
func GetRequestID(req *TRequest) string {
	if id, exists := req.Params["__request_id"]; exists {
		return id
	}
	return ""
}

// Helper function to get auth token from request
func GetAuthToken(req *TRequest) string {
	if token, exists := req.Params["__auth_token"]; exists {
		return token
	}
	return ""
}
