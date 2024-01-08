// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package networking

import "go.uber.org/zap"

// NewNetworkingEngine creates a new networking engine.
func NewNetworkingEngine(
	logger *zap.SugaredLogger, reqs ClientRequirements,
	introductoryHost, joinKey string,
) {
	// Mark the logger as being for engine.networking.
	logger = logger.Named("engine.networking")

}
