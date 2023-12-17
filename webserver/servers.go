// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package webserver

import (
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

func (w *WebServer) fasthttpServe(ln net.Listener) error {
	// Defines the fallback router. Handles mainly the frontend.
	fallbackRouter := fasthttpadaptor.NewFastHTTPHandler(w.generateHttpRoutes(false))

	// Defines the main router.
	mainRouter := router.New()

	// Add the routes required for RPC if the RPC server is not nil.
	if w.rpcServer != nil {
		mainRouter.POST("/rpc/{method}", w.rpcServer.FastHTTPHandler)
		mainRouter.GET("/rpc", w.rpcServer.FastHTTPHandler)
	}

	// Add the API routes.
	w.apiServer.AddToFasthttpRouter(mainRouter)

	// Make it fallback to the fallback router when the route is not in the main router.
	mainRouter.NotFound = fallbackRouter

	// Serve the main router.
	return fasthttp.Serve(ln, mainRouter.Handler)
}
