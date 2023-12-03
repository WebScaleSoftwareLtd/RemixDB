// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package webserver

import (
	"bytes"
	"net"
	"net/http"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func (w *WebServer) netHttpServeH2c(ln net.Listener) error {
	return http.Serve(ln, h2c.NewHandler(w.generateHttpRoutes(true), &http2.Server{}))
}

func (w *WebServer) netHttpServeTls(ln net.Listener, certFile, keyFile string) error {
	return http.ServeTLS(ln, w.generateHttpRoutes(true), certFile, keyFile)
}

var slashRpcB = []byte("/rpc")

func (w *WebServer) fasthttpServe(ln net.Listener) error {
	r := fasthttpadaptor.NewFastHTTPHandler(w.generateHttpRoutes(false))

	rpcRouter := router.New()
	fasthttpRpc := func(ctx *fasthttp.RequestCtx) {
		w.rpcServerLock.RLock()
		s := w.rpcServer
		w.rpcServerLock.RUnlock()

		s.FastHTTPHandler(ctx)
	}
	rpcRouter.POST("/rpc/{method}", fasthttpRpc)
	rpcRouter.GET("/rpc", fasthttpRpc)

	return fasthttp.Serve(ln, func(ctx *fasthttp.RequestCtx) {
		// Check if it starts with /rpc.
		path := ctx.Path()
		if len(path) >= 4 && bytes.Equal(path[:4], slashRpcB) {
			// Switch to the RPC router.
			rpcRouter.Handler(ctx)
			return
		}

		// Switch to the normal router with a compatibility layer.
		r(ctx)
	})
}
