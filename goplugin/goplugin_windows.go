//go:build windows

// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package goplugin

import (
	"sync"

	"remixdb.io/goplugin/interpretermode"
	"remixdb.io/logger"
)

// GoPluginCompiler is used to define the Go plugin compiler. This turns the specified
// Go code into a plugin that can be used by RemixDB.
type GoPluginCompiler struct {
	m  map[string]interpretermode.PluginLike
	mu sync.Mutex
}

// Compile is used to compile the Go plugin or return a cached version. It is compiled
// within the project zip specified. This is thread safe.
func (g GoPluginCompiler) Compile(code string) (Plugin, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if val, ok := g.m[code]; ok {
		return val, nil
	}

	i, err := interpretermode.Parse(code)
	if err != nil {
		return nil, err
	}

	g.m[code] = i
	return i, nil
}

// NewGoPluginCompiler is used to create a new Go plugin compiler.
func NewGoPluginCompiler(_ logger.Logger, _ string) *GoPluginCompiler {
	return &GoPluginCompiler{
		m: make(map[string]interpretermode.PluginLike),
	}
}
