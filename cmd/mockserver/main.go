// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	"remixdb.io/config"
	"remixdb.io/internal/api"
	"remixdb.io/internal/api/mockimplementation"
	"remixdb.io/internal/errhandler"
	"remixdb.io/internal/logger"
	"remixdb.io/internal/webserver"
)

func main() {
	// Defines the logger.
	l := logger.NewStdLogger()

	// Just serve locally on port 8080.
	conf := &config.ServerConfig{
		Host: "127.0.0.1:8080",
		H2C:  true,
	}
	ws := webserver.NewWebServer(webserver.WebServerConfig{
		Logger:    l,
		Config:    conf,
		RPCServer: nil,
		APIServer: api.NewServer(
			mockimplementation.New(), errhandler.Handler{
				Logger: l,
			},
		),
	})

	// Start the web server.
	if err := ws.Serve(); err != nil {
		panic(err)
	}
}
