// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"remixdb.io/errhandler"
)

type sharedRequest interface {
	// Context is used to get the context for the request.
	Context() context.Context

	// Method is used to get the method for the request.
	Method() string

	// Hostname is used to get the hostname for the request.
	Hostname() string

	// SchemaHash is used to get the schema hash for the request.
	SchemaHash() string

	// Body is used to return the body of the setup message.
	Body() []byte

	// ReturnCustomException is used to return a custom exception.
	ReturnCustomException(code int, exceptionName string, body any) error

	// ReturnRemixDBException is used to return a RemixDB exception.
	ReturnRemixDBException(httpCode int, code, message string) error

	// ReturnRemixBytes is used to return a RemixDB RPC response.
	ReturnRemixBytes(code int, data []byte)
}

type websocketRequest interface {
	sharedRequest

	// Next is used to wait for the user to ask for the next message. Returns false if the connection is closed.
	Next() bool

	// ReturnEOF is used to return an EOF error.
	ReturnEOF()
}

// PartitionHandler is used to handle a partition. Note that errors should not be used for user facing errors.
type PartitionHandler func(ctx *RequestCtx) (*Response, error)

// Server is used to define a RPC server. This server is built to be very low level.
type Server struct {
	// ErrorHandler is used to handle any errors.
	ErrorHandler errhandler.Handler

	// ListenToXForwardedFor is used to define if the server should listen to the X-Forwarded-Host header.
	ListenToXForwardedHost bool

	// PartitionsEnabled is used to define if partitions are enabled. If this is off, the partition '%' will be used.
	PartitionsEnabled bool

	// GetPartitionHandler is used to get the partition handler for a partition. If the partition does not exist, it will return
	// nil for both the handler and the error.
	GetPartitionHandler func(partition string) (PartitionHandler, error)
}

// PanicError is used to wrap a panic that is not of type error.
type PanicError struct {
	// Value is used to define the value of the panic.
	Value any
}

// Error is used to return the error.
func (p PanicError) Error() string {
	return fmt.Sprint(p.Value)
}

var nl = []byte("\n")

// Handles a request for a RPC. Supports both sharedRequest and websocketRequest.
func (s *Server) handleRpc(r sharedRequest) {
	// Handle panics.
	defer func() {
		if re := recover(); re != nil {
			// Get the error.
			err, ok := r.(error)
			if !ok {
				err = PanicError{r}
			}

			// Handle the error.
			s.ErrorHandler.HandleError(err)

			// Close the connection.
			_ = r.ReturnRemixDBException(500, "internal_server_error", "Internal server error.")
		}
	}()

	// Split the first line off the body and parse it.
	body := r.Body()
	sp := bytes.SplitN(body, nl, 2)
	if len(sp) != 2 {
		_ = r.ReturnRemixDBException(400, "invalid_request_body", "Invalid request body.")
		return
	}
	jsonLine := sp[0]
	body = sp[1]

	// Parse the JSON line.
	var authData map[string]string
	if err := json.Unmarshal(jsonLine, &authData); err != nil {
		_ = r.ReturnRemixDBException(400, "invalid_request_body", "Invalid request body.")
		return
	}

	// Get the partition.
	partitionName := "%"
	if s.PartitionsEnabled {
		partitionName = r.Hostname()
	}
	partition, err := s.GetPartitionHandler(partitionName)
	if err != nil {
		// Handle any exceptions.
		s.ErrorHandler.HandleError(err)
		_ = r.ReturnRemixDBException(500, "internal_server_error", "Internal server error.")
		return
	}

	// Handle if the partition does not exist.
	if partition == nil {
		_ = r.ReturnRemixDBException(404, "partition_does_not_exist", "The hostname does not exist as a partition.")
		return
	}

	// Handle the request context creation.
	resp, err := partition(&RequestCtx{
		Partition:  partitionName,
		Method:     r.Method(),
		AuthData:   authData,
		Context:    r.Context(),
		SchemaHash: r.SchemaHash(),
		Body:       body,
	})
	if err != nil {
		_ = r.ReturnRemixDBException(500, "internal_server_error", "Internal server error.")
		s.ErrorHandler.HandleError(err)
		return
	}

	// Handles a 204.
	if resp == nil {
		// If this is a websocket, make this a empty iterator.
		if ws, ok := r.(websocketRequest); ok {
			// Wait for the next.
			if ws.Next() {
				// Return EOF.
				ws.ReturnEOF()
			}
		}

		// Return here.
		return
	}

	// If this is a cursor, handle it.
	if resp.cursorHn != nil {
		ws, ok := r.(websocketRequest)
		if !ok {
			// Someone tried to use a cursor on a non-websocket request. Return a 400.
			_ = r.ReturnRemixDBException(400, "non_cursor_request", "This request type does not support cursors.")
			return
		}

		h := *resp.cursorHn
		for ws.Next() {
			// Handle the cursor.
			b, err := h.hn()
			if err != nil {
				// Return a error.
				_ = r.ReturnRemixDBException(500, "internal_server_error", "Internal server error.")
				s.ErrorHandler.HandleError(err)
				h.cleanup()
				return
			}

			if b == nil {
				// Return EOF.
				ws.ReturnEOF()
				h.cleanup()
				return
			}

			// Return the bytes.
			s := 200
			if len(b) == 0 {
				s = 204
			}
			ws.ReturnRemixBytes(s, b)
		}

		// Run the cleanup and return.
		h.cleanup()
		return
	}

	// Handle errors.
	if resp.err != nil {
		if resp.err.isCustom {
			// Return a custom exception.
			_ = r.ReturnCustomException(resp.err.httpCode, resp.err.codeOrType, resp.err.data)
		} else {
			// Return a RemixDB exception.
			_ = r.ReturnRemixDBException(resp.err.httpCode, resp.err.codeOrType, resp.err.data.(string))
		}
		return
	}

	// Handle if this is a websocket.
	if _, ok := r.(websocketRequest); ok {
		// Someone tried to use a cursor on a non-websocket request. Return a 400.
		_ = r.ReturnRemixDBException(400, "non_cursor_request", "This request type does not support non-cursors.")
		return
	}

	// Return the bytes.
	st := 200
	if len(resp.data) == 0 {
		st = 204
	}
	r.ReturnRemixBytes(st, resp.data)
}
