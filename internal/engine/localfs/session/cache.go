// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package session

import (
	"sync"

	"remixdb.io/ast"
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

	partitionLocks   map[string]*utils.NamedLock
	partitionLocksMu sync.RWMutex
}

// CleanPartition is used to clean the cache for a partition. Use with care! Make sure there's no sessions running for the partition.
func (c *Cache) CleanPartition(partition string) {
	c.contracts.Delete(partition)

	c.partitionLocksMu.Lock()
	if c.partitionLocks != nil {
		delete(c.partitionLocks, partition)
	}
	c.partitionLocksMu.Unlock()
}
