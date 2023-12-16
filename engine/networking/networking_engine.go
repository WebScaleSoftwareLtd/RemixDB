// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package networking

import "remixdb.io/logger"

// NewNetworkingEngine creates a new networking engine.
func NewNetworkingEngine(
	logger logger.Logger, reqs ClientRequirements,
	introductoryHost, joinKey string,
) {
	// Mark the logger as being for engine.networking.
	logger = logger.Tag("engine.networking")

}
