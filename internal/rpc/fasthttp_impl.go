// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package rpc

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

var fasthttpWsUpgrader = websocket.FastHTTPUpgrader{}

type fasthttpHandler struct {
	ctx                    *fasthttp.RequestCtx
	method                 string
	sent                   bool
	listenToXForwardedHost bool
}

func (h *fasthttpHandler) Context() context.Context {
	return h.ctx
}

func (h *fasthttpHandler) Method() string {
	return h.method
}

func (h *fasthttpHandler) SchemaHash() string {
	return string(h.ctx.Request.Header.Peek("X-RemixDB-Schema-Hash"))
}

func (h *fasthttpHandler) Hostname() string {
	if h.listenToXForwardedHost {
		s := string(h.ctx.Request.Header.Peek("X-Forwarded-Host"))
		if s != "" {
			return s
		}
	}

	return string(h.ctx.Request.Header.Host())
}

func (h *fasthttpHandler) Body() []byte {
	return h.ctx.Request.Body()
}

func (h *fasthttpHandler) ReturnCustomException(code int, exceptionName string, body any) error {
	if h.sent {
		return nil
	}
	h.sent = true
	h.ctx.Response.Header.Set("Content-Type", "application/json")
	h.ctx.Response.Header.Set("X-RemixDB-Exception", exceptionName)
	h.ctx.SetStatusCode(code)
	return json.NewEncoder(h.ctx).Encode(body)
}

func (h *fasthttpHandler) ReturnRemixDBException(httpCode int, code, message string) error {
	if h.sent {
		return nil
	}
	h.sent = true
	h.ctx.Response.Header.Set("Content-Type", "application/json")
	h.ctx.SetStatusCode(httpCode)
	return json.NewEncoder(h.ctx).Encode(map[string]string{
		"code":    code,
		"message": message,
	})
}

func (h *fasthttpHandler) ReturnRemixBytes(code int, data []byte) {
	if h.sent {
		return
	}
	h.sent = true
	h.ctx.Response.Header.Set("Content-Type", "application/remixdb-rpc")
	h.ctx.SetStatusCode(code)
	h.ctx.Write(data)
}

var _ sharedRequest = &fasthttpHandler{}

var (
	contentTypeB = []byte("application/x-remixdb-rpc-mixed")
	websocketB   = []byte("websocket")
)

// FastHTTPHandler is used to handle a HTTP request via the fasthttp package.
func (s *Server) FastHTTPHandler(ctx *fasthttp.RequestCtx) {
	// Add X-Is-RemixDB: true to the response.
	ctx.Response.Header.Set("X-Is-RemixDB", "true")

	// Get the method.
	m := ctx.UserValue("method").(string)
	if m != "" {
		// Check if this is not a POST request.
		if !ctx.IsPost() {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			_, _ = ctx.WriteString("Method not allowed")
			return
		}

		// Check if the content type is correct.
		if !bytes.Equal(ctx.Request.Header.ContentType(), contentTypeB) {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			_, _ = ctx.WriteString("Invalid content type")
			return
		}

		// Handle the request.
		h := &fasthttpHandler{
			ctx:                    ctx,
			method:                 m,
			listenToXForwardedHost: s.ListenToXForwardedHost,
		}
		s.handleRpc(h)
		if !h.sent {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
		}
		return
	}

	// Check if this is a GET request.
	if !ctx.IsGet() {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		_, _ = ctx.WriteString("Method not allowed")
		return
	}

	// Handle if there is no websocket upgrade.
	if !bytes.Equal(ctx.Request.Header.Peek("Upgrade"), websocketB) {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		_, _ = ctx.WriteString("Invalid upgrade - did you mean to connect over websocket?")
		return
	}

	// Get the host.
	var host string
	if s.ListenToXForwardedHost {
		host = string(ctx.Request.Header.Peek("X-Forwarded-Host"))
		if host != "" {
			goto postHostResolve
		}
	}
	host = string(ctx.Request.Header.Host())

postHostResolve:
	// Handle the websocket upgrade.
	_ = fasthttpWsUpgrader.Upgrade(ctx, func(ws *websocket.Conn) {
		s.handleWebsocketConn(ws, host)
	})
}

var _ fasthttp.RequestHandler = (*Server)(nil).FastHTTPHandler
