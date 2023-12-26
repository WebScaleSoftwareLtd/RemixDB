// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	"runtime"

	"remixdb.io/internal/goplugin"
	"remixdb.io/internal/logger"
	"remixdb.io/internal/utils"
)

func nonWindowsSetup(logger logger.Logger) {
	goplugin.NewGoPluginCompiler(logger, "")
}

func main() {
	// Make sure there is only one of us or RemixDB running.
	utils.EnsureSingleInstance()

	// Handle Go plugin setup.
	logger := logger.NewStdLogger()
	if runtime.GOOS != "windows" {
		nonWindowsSetup(logger)
	}
}
