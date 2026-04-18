package cache

import (
	"link-storage-service/internal/domain"
	"sync"
	"time"
)

type entry struct {
	link      *domain.Link
	expiresAt time.Time
}

type MemoryCache struct {
	mu      sync.RWMutex
	data    map[string]entry
	ttl     time.Duration
	sweep   time.Duration
}

func NewMemoryCache(ttl, sweep time.Duration) *MemoryCache {
	c := &MemoryCache{
		data:  make(map[string]entry),
		ttl:   ttl,
		sweep: sweep,
	}
	go c.cleanupLoop()
	return c
}

func (c *MemoryCache) Get(shortCode string) (*domain.Link, bool) {
	c.mu.RLock()
	e, ok := c.data[shortCode]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.link, true
}

func (c *MemoryCache) Set(shortCode string, link *domain.Link) {
	c.mu.Lock()
	c.data[shortCode] = entry{link: link, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

func (c *MemoryCache) Delete(shortCode string) {
	c.mu.Lock()
	delete(c.data, shortCode)
	c.mu.Unlock()
}

func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(c.sweep)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for k, e := range c.data {
			if now.After(e.expiresAt) {
				delete(c.data, k)
			}
		}
		c.mu.Unlock()
	}
}

