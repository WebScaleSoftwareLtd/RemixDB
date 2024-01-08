//go:build windows

// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package goplugin

import (
	"plugin"

	"go.uber.org/zap"
)

// GoPluginCompiler is used to define the Go plugin compiler. This turns the specified
// Go code into a plugin that can be used by RemixDB.
type GoPluginCompiler struct{}

// Compile is used to compile the Go plugin or return a cached version. It is compiled
// within the project zip specified. This is thread safe.
func (g GoPluginCompiler) Compile(code string) (*plugin.Plugin, error) {
	panic("not implemented on windows")
}

// NewGoPluginCompiler is used to create a new Go plugin compiler.
func NewGoPluginCompiler(logger *zap.SugaredLogger, path string) GoPluginCompiler {
	panic("not implemented on windows")
}
