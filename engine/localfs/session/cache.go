// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package session

import (
	"remixdb.io/ast"
	"remixdb.io/utils"
)

// Cache is the cache object that can be used to cache data across many sessions.
type Cache struct {
	contracts utils.TLRUCache[string, map[string]*ast.ContractToken]
}

// CleanPartition is used to clean the cache for a partition.
func (c *Cache) CleanPartition(partition string) {
	c.contracts.Delete(partition)
}
