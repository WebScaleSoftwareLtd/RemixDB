// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package rpc

import "context"

// RequestCtx is the context for a request.
type RequestCtx struct {
	context.Context

	// Partition is the partition that was sent with the request. This should not be modified.
	Partition string

	// Method is the method that was sent with the request. This should not be modified.
	Method string

	// AuthData is the authentication data that was sent with the request. This should not be modified.
	AuthData map[string]string

	// SchemaHash is the schema hash that was sent with the request. This should not be modified.
	SchemaHash string

	// Body is the body that was sent with the request. This should not be modified.
	Body []byte
}

type errResponse struct {
	httpCode   int
	codeOrType string
	data       any
	isCustom   bool
}

// Response is used to send a response.
type Response struct {
	cursorHn func() ([]byte, error)
	err      *errResponse
	data     []byte
}

// Cursor is used to return a cursor method. A cursor when both values are set to nil will return EOF.
func Cursor(hn func() ([]byte, error)) *Response { return &Response{cursorHn: hn} }

// RemixDBBytes is used to return a RemixDB RPC response.
func RemixDBBytes(data []byte) *Response { return &Response{data: data} }

// RemixDBException is used to return a RemixDB exception.
func RemixDBException(httpCode int, code, message string) *Response {
	return &Response{
		err: &errResponse{
			httpCode:   httpCode,
			codeOrType: code,
			data:       message,
		},
	}
}

// CustomException is used to return a custom exception.
func CustomException(httpCode int, exceptionName string, body any) *Response {
	return &Response{
		err: &errResponse{
			httpCode:   httpCode,
			codeOrType: exceptionName,
			data:       body,
			isCustom:   true,
		},
	}
}
