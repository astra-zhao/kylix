// session.go — Server-side session management for KylixBoot (v0.9.0 P1).
//
// A Session holds string key/value pairs on the server, identified by a
// random session ID carried in the KYLIX_SID cookie. This complements the
// stateless JWT flow: JWT for APIs, sessions for HTML forms (login state,
// CSRF tokens, flash messages, remember-me).
//
// Usage from a Kylix handler (after boot.Use(boot.Sessions())):
//
//	req.SessionSet('uid', '42');
//	name := req.SessionGet('uid');
//	req.SessionDestroy();  // logout
//
// Storage is in-memory with a mutex; expired sessions are swept lazily.
// Session values are strings to keep the Go and LLVM boot runtimes
// byte-identical in behavior.
package boot

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// SessionCookieName is the cookie carrying the session ID.
const SessionCookieName = "KYLIX_SID"

// Default TTLs (overridable via NewSessionStore options).
const (
	DefaultSessionTTL    = 24 * time.Hour
	DefaultRememberTTL   = 30 * 24 * time.Hour
	sessionSweepInterval = time.Minute
	csrfSessionKey       = "__csrf"
)

// Session is one server-side session.
type Session struct {
	ID   string
	data map[string]string

	CreatedAt time.Time
	ExpiresAt time.Time

	// New is true when this middleware run created the session; the
	// middleware then writes the Set-Cookie header on the response.
	New bool
	// Remember marks a persistent cookie (RememberTTL instead of a
	// browser-session cookie). Set via SessionMarkRemember at login.
	Remember bool
	// CookieDirty forces a Set-Cookie re-send (e.g. after Remember flip).
	CookieDirty bool
	// Destroyed marks the session deleted server-side; the middleware
	// sends an expired cookie so the browser drops it.
	Destroyed bool

	// store is set by the store/middleware so SessionDestroy can remove
	// the entry server-side.
	store *SessionStore
}

// Get returns the string value for key, or "" when absent.
func (s *Session) Get(key string) string {
	if s == nil {
		return ""
	}
	return s.data[key]
}

// Set stores a string value.
func (s *Session) Set(key, value string) {
	if s == nil {
		return
	}
	if s.data == nil {
		s.data = map[string]string{}
	}
	s.data[key] = value
}

// Delete removes a key.
func (s *Session) Delete(key string) {
	if s == nil {
		return
	}
	delete(s.data, key)
}

// SessionGet returns the session value for key ("" when no session).
func (r *Request) SessionGet(key string) string { return r.Session.Get(key) }

// SessionSet stores a session value for key.
func (r *Request) SessionSet(key, value string) { r.Session.Set(key, value) }

// SessionDelete removes a session key.
func (r *Request) SessionDelete(key string) { r.Session.Delete(key) }

// SessionMarkRemember flips the current session to a persistent
// (remember-me) cookie: 30-day expiry, re-sent on this response.
func (r *Request) SessionMarkRemember() {
	if r.Session == nil {
		return
	}
	r.Session.Remember = true
	r.Session.CookieDirty = true
}

// SessionRegenerate rotates the session ID, preserving all session values.
// Called at login to defeat session fixation: the pre-authentication session
// entry (and its ID) is discarded server-side and a fresh ID takes over —
// mirrors the LLVM @__kylix_boot_session_regen sequence (v0.10.0 P2).
func (r *Request) SessionRegenerate() {
	if r.Session == nil || r.Session.store == nil {
		return
	}
	old := r.Session
	newS := old.store.Create(old.Remember)
	for k, v := range old.data {
		newS.Set(k, v)
	}
	old.store.Delete(old.ID)
	old.Destroyed = true // the middleware drops the old cookie value
	newS.CookieDirty = true
	r.Session = newS
}

// SessionDestroy deletes the session server-side and drops the cookie.
func (r *Request) SessionDestroy() {
	if r.Session == nil {
		return
	}
	if r.Session.store != nil {
		r.Session.store.Delete(r.Session.ID)
	}
	r.Session.Destroyed = true
}

// SessionCSRFToken returns the session's CSRF token, generating and
// storing one on first call (see csrf.go for the validation middleware).
func (r *Request) SessionCSRFToken() string {
	if r.Session == nil {
		return ""
	}
	if tok := r.Session.Get(csrfSessionKey); tok != "" {
		return tok
	}
	tok := randomID(16)
	r.Session.Set(csrfSessionKey, tok)
	return tok
}

// SessionStore holds all live sessions with lazy expiry sweeping.
type SessionStore struct {
	mu          sync.Mutex
	sessions    map[string]*Session
	TTL         time.Duration // idle expiry for normal sessions
	RememberTTL time.Duration // expiry after SessionMarkRemember
	lastSweep   time.Time
}

// NewSessionStore creates a store with default TTLs.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions:    map[string]*Session{},
		TTL:         DefaultSessionTTL,
		RememberTTL: DefaultRememberTTL,
		lastSweep:   time.Now(),
	}
}

// Get returns the live session with the given ID, or nil.
func (st *SessionStore) Get(id string) *Session {
	if st == nil || id == "" {
		return nil
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	s := st.sessions[id]
	if s == nil {
		return nil
	}
	if time.Now().After(s.ExpiresAt) {
		delete(st.sessions, id)
		return nil
	}
	return s
}

// Create makes a new session. remember=true uses RememberTTL.
func (st *SessionStore) Create(remember bool) *Session {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.sweepLocked()
	now := time.Now()
	ttl := st.TTL
	if remember {
		ttl = st.RememberTTL
	}
	s := &Session{
		ID:        randomID(32),
		data:      map[string]string{},
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		Remember:  remember,
		New:       true,
	}
	s.store = st
	st.sessions[s.ID] = s
	return s
}

// Delete removes a session immediately (logout).
func (st *SessionStore) Delete(id string) {
	if st == nil || id == "" {
		return
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.sessions, id)
}

// sweepLocked drops expired sessions; called under st.mu at most once per
// sessionSweepInterval to keep O(live) cost off the hot path.
func (st *SessionStore) sweepLocked() {
	now := time.Now()
	if now.Sub(st.lastSweep) < sessionSweepInterval {
		return
	}
	st.lastSweep = now
	for id, s := range st.sessions {
		if now.After(s.ExpiresAt) {
			delete(st.sessions, id)
		}
	}
}

// Sessions middleware loads (or creates) the request's session and writes
// the session cookie when needed. Attach before any handler that touches
// req.Session or the CSRF middleware:
//
//	boot.Use(boot.Sessions())
func Sessions() Middleware {
	store := NewSessionStore()
	return SessionWithStore(store)
}

// SessionWithStore is Sessions() with an externally-owned store (share one
// store across middleware instances, or preload sessions in tests).
func SessionWithStore(store *SessionStore) Middleware {
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			isNew := false
			if req.Session == nil {
				if sid := req.Cookie(SessionCookieName); sid != "" {
					if s := store.Get(sid); s != nil {
						req.Session = s
					}
				}
				if req.Session == nil {
					req.Session = store.Create(false)
				}
				isNew = req.Session.New
				// Sliding expiry: extend on every authenticated use.
				ttl := store.TTL
				if req.Session.Remember {
					ttl = store.RememberTTL
				}
				req.Session.ExpiresAt = time.Now().Add(ttl)
				req.Session.store = store
			}
			resp := next(req)
			// Clear the stored flag after the handler ran: "new" is a
			// property of this request, not of the session — otherwise
			// every reuse would re-send Set-Cookie.
			req.Session.New = false
			if resp != nil && (isNew || req.Session.CookieDirty || req.Session.Destroyed) {
				writeSessionCookie(resp, req.Session)
			}
			return resp
		}
	}
}

// writeSessionCookie emits the Set-Cookie header for the session.
func writeSessionCookie(resp *Response, s *Session) {
	if s.Destroyed {
		resp.WithCookieAttrs(SessionCookieName, "", "Path=/; Max-Age=0; HttpOnly")
		return
	}
	attrs := "Path=/; HttpOnly"
	if s.Remember {
		attrs += "; Max-Age=" + itoaSeconds(int(DefaultRememberTTL/time.Second))
	}
	resp.WithCookieAttrs(SessionCookieName, s.ID, attrs)
}

// randomID returns a cryptographically random hex string of n bytes.
func randomID(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand never fails on the supported platforms; fall back
		// to a time-derived ID rather than panicking in a request path.
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))[:2*n]
	}
	return hex.EncodeToString(b)
}

// itoaSeconds avoids importing strconv for one call site.
func itoaSeconds(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
