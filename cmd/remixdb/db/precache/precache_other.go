//go:build !windows

// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package precache

import (
	"github.com/urfave/cli/v2"
	"remixdb.io/internal/goplugin"
	"remixdb.io/internal/logger"
)

// Precache is used to precache the Go download.
func Precache(_ *cli.Context) error {
	logger := logger.NewStdLogger()
	goplugin.NewGoPluginCompiler(logger, "")
	return nil
}
