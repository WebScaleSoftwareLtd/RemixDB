// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package main

import (
	"runtime"

	"remixdb.io/goplugin"
	"remixdb.io/internal/zipgen"
	"remixdb.io/logger"
)

func nonWindowsSetup(logger logger.Logger) {
	goplugin.NewGoPluginCompiler(logger, zipgen.CreateZip(map[string]any{
		"lol": "hi",
	}), zipgen.CreateZip(map[string]any{
		"go.mod": "module remixdb.io\n",
	}))
}

func main() {
	logger := logger.NewStdLogger()

	if runtime.GOOS != "windows" {
		nonWindowsSetup(logger)
	}
}
