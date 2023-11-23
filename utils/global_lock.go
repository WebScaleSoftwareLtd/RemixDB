// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package utils

import "sync"

type globalLock struct {
	locks map[string]*sync.Mutex
	mu    sync.Mutex
}

// GlobalLock is used to lock a global resource within the application.
var GlobalLock = globalLock{
	locks: map[string]*sync.Mutex{},
}

// Acquire acquires a lock on the given resource.
func (g *globalLock) Acquire(name string, fn func()) {
	g.mu.Lock()
	x, ok := g.locks[name]
	if !ok {
		x = &sync.Mutex{}
		g.locks[name] = x
	}
	g.mu.Unlock()

	x.Lock()
	defer x.Unlock()

	fn()
}
