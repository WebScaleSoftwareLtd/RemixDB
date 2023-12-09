// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package session

import "remixdb.io/engine"

var _ engine.StructSessionMethods = (*Session)(nil)
