//go:build !windows

// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package precache

import (
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"remixdb.io/internal/goplugin"
)

// Precache is used to precache the Go download.
func Precache(_ *cli.Context) error {
	loggerInstance, _ := zap.NewProduction()
	logger := loggerInstance.Sugar()
	defer logger.Sync()
	goplugin.NewGoPluginCompiler(logger, "")
	return nil
}
