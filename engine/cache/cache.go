// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package cache

import (
	"sync"
	"sync/atomic"
	"time"

	"remixdb.io/engine/ifaces/fsmiddleware"
)

type cacheItem struct {
	t *time.Timer
	p string
	r string
	v []byte
}

type queueItem[T any] struct {
	v T
	n *queueItem[T]
}

type queue[T comparable] struct {
	h *queueItem[T]
	t *queueItem[T]
}

func (q *queue[T]) push(v T) {
	// Create the item.
	i := &queueItem[T]{v: v}

	if q.h == nil {
		// Set the head.
		q.h = i
		q.t = i
		return
	}

	// Set the tail.
	q.t.n = i
	q.t = i
}

func (q *queue[T]) remove(v T) bool {
	// Defines the previous item.
	var prev *queueItem[T]

	// Loop through the queue.
	for i := q.h; i != nil; i = i.n {
		if i.v != v {
			// Not the item we are looking for.
			prev = i
			continue
		}

		if prev == nil {
			// This is the head. Wipe the cache.
			q.h = nil
			q.t = nil
			return true
		}

		if i.n == nil {
			// This is the tail. Move the tail back.
			q.t = prev
			prev.n = nil
			return true
		}

		// This is not a head or tail.
		prev.n = i.n
		return true
	}
	return false
}

type evictor struct {
	t uint64
	c int
	q queue[*cacheItem]
}

func (e *evictor) remItem(i *cacheItem) {
	if e.q.remove(i) {
		// Remove the item from the queue.
		e.t -= uint64(len(i.v))

		// Remove 1 from the count.
		e.c--
	}
}

var emptySlice = []*cacheItem{}

// removedItems will be nil if there's not enough space in the cache. Do
// not try and write to the slice.
func (e *evictor) prepItemSize(s, max uint64) (removedItems []*cacheItem) {
	if e.t+s <= max {
		// There's enough space in the cache.
		return emptySlice
	}

	if s > max {
		// The item is too big to fit in the cache.
		return nil
	}

	if e.q.h == e.q.t {
		// If one item is hogging the cache, remove it.
		v := e.q.h.v
		e.q.h = nil
		e.q.t = nil
		e.t = 0
		e.c = 0
		return []*cacheItem{v}
	}

	// Figure out how many items we need to remove.
	needToRemove := 0
	cacheSize := e.t + s
	qi := e.q.h
	for cacheSize > max {
		// Add to the count.
		needToRemove++

		// Subtract from the length.
		cacheSize -= uint64(len(qi.v.v))

		// Move to the next item.
		qi = qi.n
	}

	// Check if the need to remove is over 70% of the cache.
	if float64(needToRemove)/float64(e.c) >= 0.7 {
		// Return nil. We don't want a whole item to hog the cache.
		return nil
	}

	// Make a slice of items to remove.
	removedItems = make([]*cacheItem, needToRemove)
	e.c -= needToRemove
	for i := 0; i < needToRemove; i++ {
		// Add the item to the slice.
		removedItems[i] = e.q.h.v

		// Remove from the length.
		e.t -= uint64(len(e.q.h.v.v))

		// Move to the next item.
		e.q.h = e.q.h.n
	}

	// Return the slice.
	return removedItems
}

type loadMonitor struct {
	lastSecond uint64

	once uintptr
}

func (l *loadMonitor) getRps() uint64 {
	// Load the last second atomic.
	v := atomic.AddUint64(&l.lastSecond, 1) - 1
	if v != 0 {
		// Return the value.
		return v
	}

	// If v is 0, there are either no requests or we need to setup the timer.
	// In any case, there is not much load right now, so it's okay to do more
	// expensive operations.
	if atomic.CompareAndSwapUintptr(&l.once, 0, 1) {
		// We are the first to swap. Set the timer.
		ticker := time.NewTicker(time.Second)
		go func() {
			for range ticker.C {
				// Reset the last second atomic.
				atomic.StoreUint64(&l.lastSecond, 0)
			}
		}()
	}

	// Return 0.
	return 0
}

// CacheMiddleware is used to define the cache middleware.
type CacheMiddleware struct {
	mu sync.RWMutex
	e  evictor
	m  map[string]map[string]*cacheItem
	l  loadMonitor

	// Size is the size of the cache.
	Size uint64
}

func wrapItemForEviction(m *CacheMiddleware, i *cacheItem) func() {
	return func() {
		// Get the write lock.
		m.mu.Lock()
		defer m.mu.Unlock()

		// Remove the item from the queue.
		m.e.remItem(i)

		// Delete the item.
		p := m.m[i.p]
		if p == nil {
			// Return here.
			return
		}
		delete(p, i.r)
		if len(p) == 0 {
			// Delete the partition.
			delete(m.m, i.p)
		}
	}
}

func (m *CacheMiddleware) writeToCache(
	rel, partition string, partitionTtl uint64,
	b []byte,
) {
	// Re-lock the lock for writing.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Prep the cache with the item size.
	removedItems := m.e.prepItemSize(uint64(len(b)), m.Size)

	// Check if we were raced.
	if m.m != nil {
		p, ok := m.m[partition]
		if ok {
			if val := p[rel].v; val != nil {
				// We were raced. Return here.
				return
			}
		}
	}

	// Handle any removed items.
	for _, i := range removedItems {
		// Stop the timer.
		i.t.Stop()

		// Delete the item from the map.
		p := m.m[i.p]
		if p != nil {
			delete(p, i.r)
			if len(p) == 0 {
				// Delete the partition.
				delete(m.m, i.p)
			}
		}
	}

	// Create the partition if it does not exist.
	p, ok := m.m[partition]
	if !ok {
		// Create the partition.
		m.m[partition] = map[string]*cacheItem{}
	}

	// Create the item.
	i := &cacheItem{
		p: partition,
		r: rel,
		v: b,
	}
	p[rel] = i

	// Create the timer.
	i.t = time.AfterFunc(
		time.Duration(partitionTtl)*time.Second,
		wrapItemForEviction(m, i),
	)

	// Add to the queue.
	m.e.q.push(i)
}

// Resets the timer if appropriate.
func (m *CacheMiddleware) timerResetter(i *cacheItem, partitionTtl uint64) {
	// Check if the server is under medium or above load.
	if m.l.getRps() > 1000 {
		// Return here. We do not want to do expensive operations.
		return
	}

	// Get the write lock.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop the timer.
	stopped := i.t.Stop()
	if !stopped {
		// The timer was already stopped.
		return
	}

	// Create the timer.
	i.t = time.AfterFunc(
		time.Duration(partitionTtl)*time.Second,
		wrapItemForEviction(m, i),
	)
}

// ReadFile is used to try and read a file from the cache. If it can't, it
// will invoke the next middleware and then possibly insert it into the cache.
func (m *CacheMiddleware) ReadFile(
	rel, partition string, partitionTtl uint64,
	next func() ([]byte, error),
) ([]byte, error) {
	// Hold the read lock.
	m.mu.RLock()

	// Defines variables to hold the item whilst using goto.
	var item *cacheItem
	var p map[string]*cacheItem
	var ok bool

	if m.m == nil {
		// No partitions are held yet.
		goto cacheMiss
	}

	// Get the partition.
	p, ok = m.m[partition]
	if !ok {
		// Partition does not exist in the cache.
		goto cacheMiss
	}

	// Get the item then release the read lock.
	item = p[rel]
	if item == nil {
		// Item does not exist in the cache.
		goto cacheMiss
	}

	// Release the read lock.
	m.mu.RUnlock()

	// Invoke the timer resetter.
	m.timerResetter(item, partitionTtl)

	// Return the item.
	return item.v, nil

cacheMiss:
	// Release the read lock.
	m.mu.RUnlock()

	// Get the item from the next middleware.
	val, err := next()
	if err != nil {
		// Error. Return it.
		return val, err
	}

	// Write the item to the cache.
	m.writeToCache(rel, partition, partitionTtl, val)

	// Return the item.
	return val, nil
}

// DeleteFile is used to delete a file from the cache.
func (m *CacheMiddleware) DeleteFile(rel, partition string) error {
	// Hold the write lock.
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.m == nil {
		// No partitions are held yet.
		return nil
	}

	// Get the partition.
	p, ok := m.m[partition]
	if !ok {
		// Partition does not exist in the cache.
		return nil
	}

	// Get the item.
	i, ok := p[rel]
	if !ok {
		// Item does not exist in the cache.
		return nil
	}

	// Remove the item from the queue.
	m.e.remItem(i)

	// Stop the timer.
	i.t.Stop()

	// Delete the item.
	delete(p, rel)
	if len(p) == 0 {
		// Delete the partition.
		delete(m.m, partition)
	}

	// Return nil.
	return nil
}

// DeletePartition is used to delete a partition from the cache.
func (m *CacheMiddleware) DeletePartition(rel, partition string) error {
	// Hold the write lock.
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.m == nil {
		// No partitions are held yet.
		return nil
	}

	// Get the partition.
	p, ok := m.m[partition]
	if !ok {
		// Partition does not exist in the cache.
		return nil
	}

	// Loop through the items.
	for _, i := range p {
		// Remove the item from the queue.
		m.e.remItem(i)

		// Stop the timer.
		i.t.Stop()
	}

	// Delete the partition.
	delete(m.m, partition)

	// Return nil.
	return nil
}

// WriteFile is used to write a file to the cache.
func (m *CacheMiddleware) WriteFile(
	rel, partition string, partitionTtl uint64, b []byte,
) error {
	// Write the item to the cache.
	m.writeToCache(rel, partition, partitionTtl, b)

	// Return nil.
	return nil
}

// RenameFile is used to rename a file in the cache.
func (m *CacheMiddleware) RenameFile(
	oldRel, newRel, partition string, partitionTtl uint64,
) error {
	// Hold the write lock.
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.m == nil {
		// No partitions are held yet.
		return nil
	}

	// Get the partition.
	p, ok := m.m[partition]
	if !ok {
		// Partition does not exist in the cache.
		return nil
	}

	// Get the item.
	i, ok := p[oldRel]
	if !ok {
		// Item does not exist in the cache.
		return nil
	}
	i.t.Stop()

	// Delete the item.
	delete(p, oldRel)
	if len(p) == 0 {
		// Delete the partition.
		delete(m.m, partition)
	}

	// Return nil.
	return nil
}

var _ fsmiddleware.Middleware = &CacheMiddleware{}
