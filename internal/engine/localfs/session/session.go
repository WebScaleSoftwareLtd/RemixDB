// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package session

import (
	"go.uber.org/zap"
	"remixdb.io/internal/engine"
	"remixdb.io/internal/engine/localfs/acid"
)

// Session is used to implement the engine.Session interface. You must call Close on the session when you are done with it.
type Session struct {
	// Logger is used to log messages.
	Logger *zap.SugaredLogger

	// Transaction is the transaction object that can be used to perform transactions.
	Transaction *acid.Transaction

	// Cache is the cache object that can be used to cache data across many sessions.
	Cache *Cache

	// PartitionName is the name of the partition.
	PartitionName string

	// DataFolder is the data folder for the database.
	DataFolder string

	// RelativePath is the relative path to the partition.
	RelativePath string

	// SchemaWriteLock is used to define if the session is a write session for schemas.
	SchemaWriteLock bool

	// Unlocker is used to unlock the partition.
	Unlocker func()

	openObjectUnlockers map[string]map[string]func()
	openStructUnlockers map[string]func()
}

func (s *Session) getObjectUnlockersMap(structName string) map[string]func() {
	if s.openObjectUnlockers == nil {
		s.openObjectUnlockers = map[string]map[string]func(){}
	}

	unlockers, ok := s.openObjectUnlockers[structName]
	if !ok {
		unlockers = map[string]func(){}
		s.openObjectUnlockers[structName] = unlockers
	}

	return unlockers
}

func (s *Session) getStructUnlockersMap() map[string]func() {
	if s.openStructUnlockers == nil {
		s.openStructUnlockers = map[string]func(){}
	}

	return s.openStructUnlockers
}

func (s *Session) ensureWriteLock() error {
	if !s.SchemaWriteLock {
		return engine.ErrReadOnlySession
	}
	return nil
}

func (s *Session) Rollback() error {
	return s.Transaction.Rollback()
}

func (s *Session) Commit() error {
	return s.Transaction.Commit(true)
}

func (s *Session) Close() error {
	// Rollback the transaction.
	if err := s.Transaction.Rollback(); err != nil {
		if err != acid.ErrAlreadyCommitted {
			return err
		}
	}

	// Handle any open object unlockers within the session.
	if s.openObjectUnlockers != nil {
		for _, structs := range s.openObjectUnlockers {
			for _, unlocker := range structs {
				unlocker()
			}
		}
	}

	// Handle any open struct unlockers within the session.
	if s.openStructUnlockers != nil {
		for _, unlocker := range s.openStructUnlockers {
			unlocker()
		}
	}

	// Unlock the partition.
	s.Unlocker()

	// No errors!
	return nil
}

var _ engine.Session = (*Session)(nil)
