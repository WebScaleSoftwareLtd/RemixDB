// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package session

import (
	"remixdb.io/engine"
	"remixdb.io/engine/localfs/acid"
	"remixdb.io/logger"
)

// Session is used to implement the engine.Session interface. You must call Close on the session when you are done with it.
type Session struct {
	// Logger is used to log messages.
	Logger logger.Logger

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

	// Unlock the partition.
	s.Unlocker()

	// No errors!
	return nil
}

var _ engine.Session = (*Session)(nil)
