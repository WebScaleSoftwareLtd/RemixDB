// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package goplugin

import (
	"bytes"
	_ "embed"
	"plugin"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"remixdb.io/utils"
)

func TestExecutionError_Error(t *testing.T) {
	err := ExecutionError{
		exitCode: 6969,
		data:     []byte("hello world! This is a test!"),
	}
	assert.Equal(t, "execution error: status 6969: hello world! This is a test!", err.Error())
}

type fuse struct {
	blown bool
}

func (f *fuse) BlowFuse() {
	f.blown = true
}

func TestGoPluginCompiler_Compile(t *testing.T) {
	// Run this test in parallel.
	t.Parallel()

	// Define the sub-tests.
	tests := []struct {
		name string

		projectFiles []byte

		goCode        string
		resultHandler func(t *testing.T, p *plugin.Plugin, err error)
	}{
		{
			name: "no error",
			projectFiles: utils.CreateZipFromMap(map[string]any{
				"go.mod": "module remixdb.io",
			}),
			goCode: `package main

type privateInterface interface {
	BlowFuse()
}

type FuseBlower interface {
	privateInterface
}

func BlowFuse(b FuseBlower) {
	b.BlowFuse()
}`,
			resultHandler: func(t *testing.T, p *plugin.Plugin, err error) {
				require.NoError(t, err)
				require.NotNil(t, p)

				f, err := p.Lookup("BlowFuse")
				require.NoError(t, err)
				require.NotNil(t, f)

				fuseInstance := &fuse{}
				reflect.ValueOf(f).Call([]reflect.Value{reflect.ValueOf(fuseInstance)})
				assert.True(t, fuseInstance.blown)
			},
		},
		{
			name: "error",
			projectFiles: utils.CreateZipFromMap(map[string]any{
				"go.mod": "module remixdb.io",
				"helloworld": map[string]any{
					"helloworld.go": `package helloworld

func HelloWorld() string {
	return "Hello World!"
}`,
				},
			}),
			goCode: `package main

import "remixdb.io/helloworld"

func HelloWorld() string {
	return helloworld.HelloWorld
}`,
			resultHandler: func(t *testing.T, p *plugin.Plugin, err error) {
				if assert.Error(t, err) {
					x, ok := err.(ExecutionError)
					if !ok {
						t.Error("error is not of type ExecutionError")
						return
					}

					assert.Equal(t, x.exitCode, 1)
					x.data = bytes.TrimSpace(x.data)
					if !bytes.HasSuffix(x.data, []byte(".go:6:9: cannot use helloworld.HelloWorld (value of type func() string) as string value in return statement")) {
						t.Errorf("unexpected error data: %s", x.data)
					}
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Run this test in parallel.
			t.Parallel()

			// Create the Go plugin compiler.
			projectZip := tt.projectFiles
			if projectZip == nil {
				projectZip = utils.CreateZipFromMap(nil)
			}
			compiler := SetupGoCompilerForTesting(t, projectZip)

			// Attempt to compile the code.
			p, err := compiler.Compile(tt.goCode)
			tt.resultHandler(t, p, err)
		})
	}
}
