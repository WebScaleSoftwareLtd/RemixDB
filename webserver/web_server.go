// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package webserver

import (
	"net"

	"remixdb.io/engine"
	"remixdb.io/rpc"
)

// WebServer is used to define a web server. Use NewWebServer to create a new instance.
type WebServer struct {
	conf Config

	engine    engine.Engine
	rpcServer *rpc.Server
}

// Serve is used to serve the web server.
func (w *WebServer) Serve() error {
	// Create the listener.
	ln, err := net.Listen("tcp", w.conf.Host)
	if err != nil {
		return err
	}

	// Handle if we should use fasthttp.
	if w.conf.HTTPSOptions == nil && !w.conf.H2C {
		return w.fasthttpServe(ln)
	}

	// Handle if we should use HTTPS.
	if w.conf.HTTPSOptions != nil {
		return w.netHttpServeTls(
			ln, w.conf.HTTPSOptions.CertFile, w.conf.HTTPSOptions.KeyFile,
		)
	}

	// Handle if we should use H2C.
	return w.netHttpServeH2c(ln)
}

// NewWebServer is used to create a new web server.
func NewWebServer(conf Config, engine engine.Engine, rpcServer *rpc.Server) *WebServer {
	return &WebServer{
		conf:      conf,
		engine:    engine,
		rpcServer: rpcServer,
	}
}
