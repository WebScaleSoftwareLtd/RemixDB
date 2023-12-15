// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package utils

import "sync"

type lock struct {
	mu    sync.RWMutex
	users int
}

// NamedLock is a lock which can be acquired by name and is safe for concurrent use.
// The handler automatically frees the lock when nobody is using it.
type NamedLock struct {
	locks   map[string]*lock
	locksMu sync.Mutex
}

// Lock is used to acquire a lock on the given resource.
func (n *NamedLock) Lock(name string) {
	n.locksMu.Lock()
	l, ok := n.locks[name]
	if !ok {
		l = &lock{}
		n.locks[name] = l
	}
	l.users++
	n.locksMu.Unlock()
	l.mu.Lock()
}

// RLock is used to acquire a read lock on the given resource.
func (n *NamedLock) RLock(name string) {
	n.locksMu.Lock()
	l, ok := n.locks[name]
	if !ok {
		l = &lock{}
		n.locks[name] = l
	}
	l.users++
	n.locksMu.Unlock()
	l.mu.RLock()
}

// Unlock is used to release a lock on the given resource.
func (n *NamedLock) Unlock(name string) {
	n.locksMu.Lock()
	l, ok := n.locks[name]
	if !ok {
		l = &lock{}
		n.locks[name] = l
	}
	l.users--
	if l.users == 0 {
		delete(n.locks, name)
	}
	n.locksMu.Unlock()
	l.mu.Unlock()
}

// RUnlock is used to release a read lock on the given resource.
func (n *NamedLock) RUnlock(name string) {
	n.locksMu.Lock()
	l, ok := n.locks[name]
	if !ok {
		l = &lock{}
		n.locks[name] = l
	}
	l.users--
	if l.users == 0 {
		delete(n.locks, name)
	}
	n.locksMu.Unlock()
	l.mu.RUnlock()
}
