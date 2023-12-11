// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package requesthandler

import (
	"context"

	"remixdb.io/engine"
	"remixdb.io/rpc"
)

type pluginFriendlyRpc struct {
	engine.Session

	req   *rpc.RequestCtx
	perms []string
	resp  *rpc.Response
}

// Permissions is used to return the permissions fetched during authentication.
func (r pluginFriendlyRpc) Permissions() []string { return r.perms }

// Context is used to return the context from RequestCtx.
func (r pluginFriendlyRpc) Context() context.Context { return r.req.Context }

// Body is used to return the body from RequestCtx.
func (r pluginFriendlyRpc) Body() []byte { return r.req.Body }

// RespondWithCursor is used to respond with a cursor. If this isn't the first usage, it will replace the previous response.
func (r *pluginFriendlyRpc) RespondWithCursor(hn func() ([]byte, error)) {
	r.resp = rpc.Cursor(hn, func() { _ = r.Close() })
}

// RespondWithRemixDBBytes is used to respond with RemixDB bytes. If this isn't the first usage, it will replace the previous response.
func (r *pluginFriendlyRpc) RespondWithRemixDBBytes(data []byte) { r.resp = rpc.RemixDBBytes(data) }

// RespondWithRemixDBException is used to respond with a RemixDB exception. If this isn't the first usage, it will replace the previous response.
func (r *pluginFriendlyRpc) RespondWithRemixDBException(httpCode int, code, message string) {
	r.resp = rpc.RemixDBException(httpCode, code, message)
}

// RespondWithCustomException is used to respond with a custom exception. If this isn't the first usage, it will replace the previous response.
func (r *pluginFriendlyRpc) RespondWithCustomException(code int, exceptionName string, body any) {
	r.resp = rpc.CustomException(code, exceptionName, body)
}
