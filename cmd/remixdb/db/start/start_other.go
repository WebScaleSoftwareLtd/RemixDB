//go:build !windows

// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package start

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"remixdb.io"
	"remixdb.io/config"
	"remixdb.io/internal/api"
	"remixdb.io/internal/api/mockimplementation"
	"remixdb.io/internal/compiler"
	"remixdb.io/internal/engine/localfs"
	"remixdb.io/internal/errhandler"
	"remixdb.io/internal/goplugin"
	"remixdb.io/internal/logger"
	"remixdb.io/internal/rpc"
	"remixdb.io/internal/rpc/requesthandler"
	"remixdb.io/internal/webserver"
)

func getConfigPath() string {
	e := os.Getenv("REMIXDB_CONFIG_PATH")
	if e != "" {
		return e
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homedir, ".remixdb", "config.yml")
}

// Start is used to start the RemixDB database.
func Start(_ *cli.Context) error {
	// Display the splash screen.
	printSplashScreen()

	// Get the configuration.
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Make sure the folder it is in exists.
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf("error creating folder to hold configuration: %w", err)
		}

		// Write the default configuration.
		err := os.WriteFile(configPath, []byte(remixdb.ExampleConfig), 0644)
		if err != nil {
			return fmt.Errorf("error writing default configuration: %w", err)
		}
	}

	// Read the configuration.
	config, err := config.Setup(configPath)
	if err != nil {
		return fmt.Errorf("error reading configuration: %w", err)
	}

	// Setup the logger.
	logger := logger.NewStdLogger()

	// Setup the Go plugin compiler.
	pluginCompiler := goplugin.NewGoPluginCompiler(logger, config.Path.GoPlugin)

	// Setup the error handler.
	errHandler := errhandler.Handler{Logger: logger}

	// Setup the compiler.
	compiler := &compiler.Compiler{GoPluginCompiler: pluginCompiler}

	// Setup the engine.
	// TODO: make this switch to the sharded version
	engine := localfs.New(logger, config.Path.Data)

	// Setup the API implementation.
	// TODO: make this the actual API implementation
	apiImpl := mockimplementation.New()

	// Setup the API server.
	apiServer := api.NewServer(apiImpl, errHandler)

	// Setup the RPC server.
	rpcServer := &rpc.Server{
		ErrorHandler:           errHandler,
		ListenToXForwardedHost: config.Server.XForwardedHost,
		PartitionsEnabled:      config.Database.PartitionsEnabled,
		GetPartitionHandler: (requesthandler.Handler{
			Engine:   engine,
			Compiler: compiler,
		}).Handle,
	}

	// Start the web server.
	return webserver.NewWebServer(webserver.WebServerConfig{
		Logger:    logger,
		Config:    config.Server,
		RPCServer: rpcServer,
		APIServer: apiServer,
	}).Serve()
}
