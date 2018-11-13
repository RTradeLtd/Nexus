package network

import (
	"runtime"
	"sync"
	"time"
)

type item struct {
	expire int64
}

type cache struct {
	dur   time.Duration
	store map[string]item
	stop  chan bool

	mux sync.RWMutex
}

// newCache creates a cache
func newCache(expire, cleanInterval time.Duration) *cache {
	c := &cache{
		dur:   expire,
		store: make(map[string]item),
		stop:  make(chan bool),
	}

	// run cleaner and set cleaner to stop when cache is destroyed
	go c.cleaner(cleanInterval)
	runtime.SetFinalizer(c, stopCacheCleaner)

	return c
}

// Cache stores given key
func (c *cache) Cache(key string) {
	c.mux.Lock()
	c.store[key] = item{time.Now().Add(c.dur).UnixNano()}
	c.mux.Unlock()
}

// Exists returns true if key exists
func (c *cache) Exists(key string) bool {
	c.mux.RLock()
	i, found := c.store[key]
	if !found {
		c.mux.RUnlock()
		return false
	}
	if i.expire > 0 && time.Now().UnixNano() > i.expire {
		c.mux.RUnlock()
		return false
	}
	c.mux.RUnlock()
	return true
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

// stopCacheCleaner is a "destructor" for cache, called by the finalizer
func stopCacheCleaner(c *cache) {
	c.stop <- true
}
