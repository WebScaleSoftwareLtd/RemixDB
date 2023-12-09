// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	"runtime"

	"remixdb.io/goplugin"
	"remixdb.io/internal/singleinstance"
	"remixdb.io/internal/zipgen"
	"remixdb.io/logger"
)

func nonWindowsSetup(logger logger.Logger) {
	goplugin.NewGoPluginCompiler(logger, "", zipgen.CreateZipFromMap(map[string]any{
		"lol": "hi",
	}))
}

func main() {
	// Make sure there is only one of us or RemixDB running.
	singleinstance.EnsureSingleInstance()

	// Handle Go plugin setup.
	logger := logger.NewStdLogger()
	if runtime.GOOS != "windows" {
		nonWindowsSetup(logger)
	}
}
