// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package goplugin

import (
	"plugin"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"remixdb.io/internal/zipgen"
)

func TestGoPluginCompiler_Compile(t *testing.T) {
	tests := []struct {
		name string

		cacheFiles   map[string]any
		projectFiles map[string]any

		goCode        string
		resultHandler func(t *testing.T, p *plugin.Plugin, err error)
	}{
		{
			name: "no cache",
			projectFiles: map[string]any{
				"go.mod": "module remixdb.io",
				"helloworld": map[string]any{
					"helloworld.go": `package helloworld

func HelloWorld() string {
	return "Hello World!"
}`,
				},
			},
			goCode: `package main

import "remixdb.io/helloworld"

func HelloWorld() string {
	return helloworld.HelloWorld()
}`,
			resultHandler: func(t *testing.T, p *plugin.Plugin, err error) {
				require.NoError(t, err)
				require.NotNil(t, p)

				f, err := p.Lookup("HelloWorld")
				require.NoError(t, err)
				require.NotNil(t, f)
				assert.Equal(t, "Hello World!", f.(func() string)())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the Go plugin compiler.
			cacheZip := zipgen.CreateZip(tt.cacheFiles)
			projectZip := zipgen.CreateZip(tt.projectFiles)
			compiler := SetupGoCompilerForTesting(t, cacheZip, projectZip)

			// Attempt to compile the code.
			p, err := compiler.Compile(tt.goCode)
			tt.resultHandler(t, p, err)
		})
	}
}
