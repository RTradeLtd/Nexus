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

func newCache(expire, cleanInterval time.Duration) *cache {
	c := &cache{
		dur:   expire,
		store: make(map[string]item),
		stop:  make(chan bool),
	}
	go c.cleaner(cleanInterval)
	runtime.SetFinalizer(c, stopCleaner)

	return c
}

func (c *cache) Cache(key string) {
	c.mux.Lock()
	c.store[key] = item{time.Now().Add(c.dur).UnixNano()}
	c.mux.Unlock()
}

func (c *cache) Get(key string) (*item, bool) {
	c.mux.RLock()
	i, found := c.store[key]
	if !found {
		c.mux.RUnlock()
		return nil, false
	}
	if i.expire > 0 && time.Now().UnixNano() > i.expire {
		c.mux.RUnlock()
		return nil, false
	}
	c.mux.RUnlock()
	return &i, true
}

func (c *cache) Size() int {
	c.mux.RLock()
	n := len(c.store)
	c.mux.RUnlock()
	return n
}

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

func stopCleaner(c *cache) {
	c.stop <- true
}
