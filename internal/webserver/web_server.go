// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package webserver

import (
	"net"

	"remixdb.io/config"
	"remixdb.io/internal/api"
	"remixdb.io/internal/logger"
	"remixdb.io/internal/rpc"
)

// WebServer is used to define a web server. Use NewWebServer to create a new instance.
type WebServer struct {
	logger logger.Logger
	conf   *config.ServerConfig

	rpcServer *rpc.Server
	apiServer api.Server
}

// Serve is used to serve the web server.
func (w *WebServer) Serve() error {
	// Create the listener.
	ln, err := net.Listen("tcp", w.conf.Host)
	if err != nil {
		return err
	}

	// Log that we bound to the port.
	w.logger.Info("Bound to "+w.conf.Host, nil)

	// Handle if we should use fasthttp.
	if w.conf.SSLCertFile == "" && !w.conf.H2C {
		return w.fasthttpServe(ln)
	}

	// Handle if we should use HTTPS.
	if w.conf.SSLCertFile != "" {
		return w.netHttpServeTls(
			ln, w.conf.SSLCertFile, w.conf.SSLKeyFile,
		)
	}

	// Handle if we should use H2C.
	return w.netHttpServeH2c(ln)
}

// WebServerConfig is used to define the web server configuration.
type WebServerConfig struct {
	// Logger is used to define the logger.
	Logger logger.Logger

	// Config is used to define the server configuration.
	Config *config.ServerConfig

	// RPCServer is used to define the RPC server.
	RPCServer *rpc.Server

	// APIServer is used to define the API server.
	APIServer api.Server
}

// NewWebServer is used to create a new web server.
func NewWebServer(conf WebServerConfig) *WebServer {
	return &WebServer{
		logger:    conf.Logger.Tag("webserver"),
		conf:      conf.Config,
		rpcServer: conf.RPCServer,
		apiServer: conf.APIServer,
	}
}
