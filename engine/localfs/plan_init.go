// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package localfs

import (
	"os"
	"path/filepath"
	"sync"

	"remixdb.io/engine/ifaces/fsmiddleware"
	"remixdb.io/engine/ifaces/planner"
)

// Engine is the local filesystem engine.
type Engine struct {
	fp string
	mu sync.RWMutex
	m  fsmiddleware.Middleware
}

// NewEngine creates a new local filesystem engine.
func NewEngine(fp string, middleware fsmiddleware.Middleware) *Engine {
	// Make sure the path exists.
	if err := os.MkdirAll(fp, 0755); err != nil {
		panic(err)
	}

	// Acquire the filesystem lock.
	mustAcquireFilesystemLock(filepath.Join(fp, "remixdb.lock"))

	// Handle any failed transactions from a crash during the last boot.
	handleFailedTransactions(fp)

	// Return the engine.
	return &Engine{fp: fp, m: middleware}
}

// AcquirePlannerLock acquires the planner lock.
func (e *Engine) AcquirePlannerLock() planner.PlanHandler {
	e.mu.Lock()
	return &planHandler{
		e:         e,
		writeLock: true,
	}
}

// AcquirePlannerReadLock acquires the planner read lock.
func (e *Engine) AcquirePlannerReadLock() planner.PlanHandler {
	e.mu.RLock()
	return &planHandler{
		e:         e,
		writeLock: false,
	}
}

var _ planner.Planner = &Engine{}

type planHandler struct {
	e         *Engine
	writeLock bool

	structs []planner.Struct
}

func (ph *planHandler) Unlock() {
	if ph.writeLock {
		ph.e.mu.Unlock()
	} else {
		ph.e.mu.RUnlock()
	}
}

var _ planner.PlanHandler = &planHandler{}
