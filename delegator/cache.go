package delegator

import (
	"net/http/httputil"
	"sync"
	"time"
)

type proxy struct {
	expire  int64
	handler *httputil.ReverseProxy
}

type cache struct {
	dur   time.Duration
	store map[string]proxy
	stop  chan bool

	mux sync.RWMutex
}

// newCache creates a cache
func newCache(expire, cleanInterval time.Duration) *cache {
	c := &cache{
		dur:   expire,
		store: make(map[string]proxy),
		stop:  make(chan bool),
	}

	// run cleaner
	go c.cleaner(cleanInterval)

	return c
}

// Cache stores given key
func (c *cache) Cache(key string, handler *httputil.ReverseProxy) {
	c.mux.Lock()
	c.store[key] = proxy{time.Now().Add(c.dur).UnixNano(), handler}
	c.mux.Unlock()
}

// Exists returns true if key exists
func (c *cache) Get(key string) (handler *httputil.ReverseProxy) {
	c.mux.RLock()
	i, found := c.store[key]
	if !found {
		c.mux.RUnlock()
		return
	}
	if i.expire > 0 && time.Now().UnixNano() > i.expire {
		c.mux.RUnlock()
		return
	}
	handler = i.handler
	c.mux.RUnlock()
	return
}

// Size returns the number of elements in the cache
func (c *cache) Size() int {
	c.mux.RLock()
	n := len(c.store)
	c.mux.RUnlock()
	return n
}

// prune removes all expired items
func (c *cache) prune() {
	c.mux.Lock()
	now := time.Now().UnixNano()
	for k, v := range c.store {
		if v.expire > 0 && now > v.expire {
			delete(c.store, k)
		}
	}
	c.mux.Unlock()
}

// cleaner runs at provided intervals to prune the store of expired items
func (c *cache) cleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			c.prune()
		case <-c.stop:
			ticker.Stop()
			return
		}
	}
}
