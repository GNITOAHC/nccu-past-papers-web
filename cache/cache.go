package cache

import (
	"sync"
	"time"
)

type item[V any] struct {
	value      V
	expiration time.Time
} // Single Item type

type Cache[K comparable, V any] struct {
	items map[K]item[V] // map of items
	mu    sync.Mutex
} // Cache type

func (i item[V]) isExpired() bool {
	return i.expiration.Before(time.Now())
}

func New[K comparable, V any]() *Cache[K, V] {
	c := &Cache[K, V]{
		items: make(map[K]item[V]),
	}

	go func() {
		for range time.Tick(24 * time.Hour) { // Check every day
			c.mu.Lock()
			for k, v := range c.items {
				if v.isExpired() {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		}
	}()

	return c
}

func (c *Cache[K, V]) Set(key K, val V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = item[V]{
		value:      val,
		expiration: time.Now().Add(ttl),
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.items[key]
	if !ok {
		return item.value, false // Zero value
	}
	if item.isExpired() {
		delete(c.items, key)
		return item.value, false
	}
	return item.value, true
}

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *Cache[K, V]) Pop(key K) (V, bool) {
	val, has := c.Get(key)
	if !has {
		return val, has
	}
	c.Delete(key)
	return val, has
}
