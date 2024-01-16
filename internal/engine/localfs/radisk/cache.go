// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package radisk

import (
	"sync"
	"time"
)

type timer struct {
	data  []byte
	timer *time.Timer
}

// Cache is used to define the interface for the cache.
type Cache struct {
	// FS is the filesystem that the cache is using.
	FS Filesystem

	m  map[uint64]*timer
	mu sync.Mutex
}

func (c *Cache) GetPage(page uint64) ([]byte, error) {
	// Try to get the page from the cache.
	c.mu.Lock()
	if c.m == nil {
		c.m = map[uint64]*timer{}
	}
	if t, ok := c.m[page]; ok {
		// Reset the timer and unlock.
		t.timer.Reset(time.Minute * 10)
		c.mu.Unlock()

		// Return the data.
		return t.data, nil
	}
	c.mu.Unlock()

	// Get the page from the filesystem.
	data, err := c.FS.GetPage(page)
	if err != nil {
		return nil, err
	}

	// Re-lock and set the page in the cache.
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.m[page]; !ok {
		// Create the timer object.
		t := &timer{data: data}

		// Start the timer.
		t.timer = time.AfterFunc(time.Minute*10, func() {
			c.mu.Lock()
			x := c.m[page]
			if x == t {
				delete(c.m, page)
			}
			c.mu.Unlock()
		})

		// Add to the map.
		c.m[page] = t
	}

	// Return the data.
	return data, nil
}

func (c *Cache) SetPage(page uint64, data []byte) error {
	// Update the page in the cache.
	c.mu.Lock()
	p := c.m[page]
	if p != nil {
		p.timer.Stop()
	}
	p = &timer{data: data}
	p.timer = time.AfterFunc(time.Minute*10, func() {
		c.mu.Lock()
		x := c.m[page]
		if x == p {
			delete(c.m, page)
		}
		c.mu.Unlock()
	})
	c.m[page] = p
	c.mu.Unlock()

	// Set the page in the filesystem.
	return c.FS.SetPage(page, data)
}

var _ Filesystem = (*Cache)(nil)
