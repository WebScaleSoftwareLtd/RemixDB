// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package session

import (
	"sync"

	"remixdb.io/internal/ast"
	"remixdb.io/internal/utils"
)

type possibleRename struct {
	R *string
	S []*ast.StructToken
}

// Cache is the cache object that can be used to cache data across many sessions.
type Cache struct {
	contracts utils.TLRUCache[string, map[string]*ast.ContractToken]
	structs   utils.TLRUCache[string, map[string]possibleRename]

	objectLocks   map[string]*utils.NamedLock
	objectLocksMu sync.RWMutex
}

// CleanPartition is used to clean the cache for a partition.
func (c *Cache) CleanPartition(partition string) {
	c.contracts.Delete(partition)

	c.objectLocksMu.Lock()
	if c.objectLocks != nil {
		delete(c.objectLocks, partition)
	}
	c.objectLocksMu.Unlock()
}
