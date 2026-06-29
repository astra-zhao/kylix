// cache.go — Thread-safe in-memory LRU cache with optional TTL.
//
// Classic LRU via container/list + map. Get/Put/Set/Delete are O(1).
// Optional per-entry TTL (milliseconds); expired entries are evicted lazily
// on access and via a background sweep when Sweep is called.
package stdlib

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// lruEntry is one cached value with its expiration time (zero = no TTL).
type lruEntry struct {
	key   string
	value interface{}
	expiresAt time.Time // zero value = never expires
}

// TCache is a thread-safe LRU cache with capacity and optional TTL.
type TCache struct {
	mu       sync.Mutex
	capacity int
	ttl      time.Duration // default TTL; 0 = no expiry
	items    map[string]*list.Element
	order    *list.List // front = most recently used
}

// NewCache creates an LRU cache with the given capacity and default TTL in
// milliseconds. ttlMs <= 0 means entries never expire by default.
func NewCache(capacity int, ttlMs int) *TCache {
	if capacity < 1 {
		capacity = 16
	}
	return &TCache{
		capacity: capacity,
		ttl:      time.Duration(ttlMs) * time.Millisecond,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Put inserts or updates a key using the cache's default TTL.
func (c *TCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.putLocked(key, value, c.ttl)
}

// PutWithTTL inserts or updates a key with a specific TTL in milliseconds.
// ttlMs <= 0 means the entry never expires.
func (c *TCache) PutWithTTL(key string, value interface{}, ttlMs int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var ttl time.Duration
	if ttlMs > 0 {
		ttl = time.Duration(ttlMs) * time.Millisecond
	}
	c.putLocked(key, value, ttl)
}

func (c *TCache) putLocked(key string, value interface{}, ttl time.Duration) {
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		entry := elem.Value.(*lruEntry)
		entry.value = value
		entry.expiresAt = c.expiryFor(ttl)
		return
	}
	entry := &lruEntry{
		key:       key,
		value:     value,
		expiresAt: c.expiryFor(ttl),
	}
	c.items[key] = c.order.PushFront(entry)
	// Evict least-recently-used if over capacity.
	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.removeElement(oldest)
		}
	}
}

func (c *TCache) expiryFor(ttl time.Duration) time.Time {
	if ttl <= 0 {
		return time.Time{}
	}
	return time.Now().Add(ttl)
}

// Get returns the value for key and a found flag. Expired entries are evicted.
func (c *TCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	elem, ok := c.items[key]
	if !ok {
		return nil, false
	}
	entry := elem.Value.(*lruEntry)
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.removeElement(elem)
		return nil, false
	}
	c.order.MoveToFront(elem)
	return entry.value, true
}

// GetString returns the cached value as a string, or "" if missing/expired.
func (c *TCache) GetString(key string) string {
	v, ok := c.Get(key)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// Delete removes a key. No-op if the key is absent.
func (c *TCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.removeElement(elem)
	}
}

// Has reports whether a non-expired entry exists for key.
func (c *TCache) Has(key string) bool {
	_, ok := c.Get(key)
	return ok
}

// Size returns the current number of entries (including any not-yet-swept
// expired ones).
func (c *TCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

// Clear removes all entries.
func (c *TCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element)
	c.order.Init()
}

// Sweep scans all entries and removes expired ones. Call periodically to
// reclaim memory from keys that are never read again.
func (c *TCache) Sweep() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	removed := 0
	for _, elem := range c.items {
		entry := elem.Value.(*lruEntry)
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			c.removeElement(elem)
			removed++
		}
	}
	return removed
}

func (c *TCache) removeElement(elem *list.Element) {
	entry := elem.Value.(*lruEntry)
	delete(c.items, entry.key)
	c.order.Remove(elem)
}
