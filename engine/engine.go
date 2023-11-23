// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package engine

import (
	"remixdb.io/engine/cache"
	"remixdb.io/engine/ifaces/planner"
	"remixdb.io/engine/localfs"
)

// NewLocalFSEngine creates a new local filesystem engine.
func NewLocalFSEngine(fp string, cacheSize uint64) planner.Planner {
	// Create the localfs engine with the cache middleware.
	engine := localfs.NewEngine(fp, &cache.CacheMiddleware{
		Size: cacheSize,
	})

	// Return the engine.
	return engine
}
