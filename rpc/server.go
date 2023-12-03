// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package rpc

import (
	"io"

	"remixdb.io/rpc/structure"
)

type httpRequest interface {
	// Method is used to get the method for the request.
	Method() string

	// SchemaHash is used to get the schema hash for the request.
	SchemaHash() string

	// Body returns a io.ReadCloser that contains the body of the request.
	Body() io.ReadCloser

	// ReturnCustomException is used to return a custom exception.
	ReturnCustomException(code int, exceptionName string, body any) error

	// ReturnRemixDBException is used to return a RemixDB exception.
	ReturnRemixDBException(httpCode int, code, message string) error

	// ReturnRemixBytes is used to return a RemixDB RPC response.
	ReturnRemixBytes(code int, data []byte)
}

type websocketRequest interface {
	// Next is used to wait for the user to ask for the next message. Returns false
	// if the connection is closed.
	Next() bool

	// Method is used to get the method for the request.
	Method() string

	// SchemaHash is used to get the schema hash for the request.
	SchemaHash() string

	// Body is used to return the body of the setup message. This starts where HTTP
	// would.
	Body() []byte

	// ReturnCustomException is used to return a custom exception.
	ReturnCustomException(code int, exceptionName string, body any) error

	// ReturnRemixDBException is used to return a RemixDB exception.
	ReturnRemixDBException(httpCode int, code, message string) error

	// ReturnRemixBytes is used to return a RemixDB RPC response.
	ReturnRemixBytes(code int, data []byte)
}

// Server is used to define a RPC server.
type Server struct {
	// Base is used to define the base structure.
	Base *structure.Base
}

// Handles a HTTP request for a RPC. This would be a non-cursor request.
func (s *Server) handleHttpRpc(r httpRequest) {
	// TODO
}

// Handles a WebSocket request for a RPC. This would be a cursor request.
func (s *Server) handleWebsocketRpc(r websocketRequest) {
	// TODO
}
