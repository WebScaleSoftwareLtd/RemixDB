// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	"runtime"

	"go.uber.org/zap"
	"remixdb.io/internal/goplugin"
	"remixdb.io/internal/utils"
)

func nonWindowsSetup(logger *zap.SugaredLogger) {
	goplugin.NewGoPluginCompiler(logger, "")
}

func main() {
	// Make sure there is only one of us or RemixDB running.
	utils.EnsureSingleInstance()

	// Handle Go plugin setup.
	loggerInstance, _ := zap.NewProduction()
	logger := loggerInstance.Sugar()
	defer logger.Sync()
	if runtime.GOOS != "windows" {
		nonWindowsSetup(logger)
	}
}
