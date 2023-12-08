// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package session

import (
	"remixdb.io/engine"
	"remixdb.io/logger"
)

// Session is used to implement the engine.Session interface. You must call Close on the session when you are done with it.
type Session struct {
	// Logger is used to log messages.
	Logger logger.Logger

	// Path is the path to the partition.
	Path string

	// WriteLock is used to define if the session is a write session.
	WriteLock bool

	// Unlocker is used to unlock the partition.
	Unlocker func()
}

var _ engine.Session = (*Session)(nil)
