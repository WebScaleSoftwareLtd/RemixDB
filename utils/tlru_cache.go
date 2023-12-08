// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package utils

import (
	"sync"
	"time"
)

type cacheItem[T any] struct {
	value T
	t     *time.Timer
}

// TLRUCache is a thread-safe TLRU cache.
type TLRUCache[K comparable, T any] struct {
	mu    sync.Mutex
	items map[K]cacheItem[T]
}

// Get returns the value for the given key.
func (t *TLRUCache[K, T]) Get(key K) (value T, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.items == nil {
		return
	}

	item, ok := t.items[key]
	if !ok {
		return
	}

	item.t.Reset(10 * time.Minute)
	return item.value, true
}

// Set sets the value for the given key.
func (t *TLRUCache[K, T]) Set(key K, value T) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.items == nil {
		t.items = map[K]cacheItem[T]{}
	}

	item, ok := t.items[key]
	if ok {
		item.t.Reset(10 * time.Minute)
		item.value = value
		return
	}

	item.t = time.AfterFunc(10*time.Minute, func() {
		t.mu.Lock()
		defer t.mu.Unlock()

		delete(t.items, key)
	})

	item.value = value
	t.items[key] = item
}

// Delete deletes the value for the given key.
func (t *TLRUCache[K, T]) Delete(key K) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.items == nil {
		return
	}

	item, ok := t.items[key]
	if !ok {
		return
	}

	item.t.Stop()
	delete(t.items, key)
}
