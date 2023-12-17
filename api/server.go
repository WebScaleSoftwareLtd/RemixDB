// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package api

import "remixdb.io/errhandler"

// Server is the structure used to implement the API.
type Server struct {
	impl       APIImplementation
	errHandler errhandler.Handler
}

// NewServer returns a new Server.
func NewServer(impl APIImplementation, errHandler errhandler.Handler) Server {
	return Server{
		impl:       impl,
		errHandler: errHandler,
	}
}
