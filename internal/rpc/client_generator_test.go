// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package rpc_test

import (
	"testing"

	"github.com/jimeh/go-golden"
	"github.com/stretchr/testify/assert"
	"remixdb.io/internal/rpc"
	"remixdb.io/internal/rpc/structure"
)

var filledStructure = &structure.Base{
	AuthenticationKeys: []string{"long_key", "key2"},
	Structs: map[string]structure.Struct{
		"OneField": {
			Comment: "used to test a single field",
			Fields: map[string]structure.StructField{
				"field": {
					Comment: "used to test a field",
					Type:    "string",
				},
			},
		},
		"ErrorWithMessageField": {
			Comment:   "used to test a error with a message field",
			Exception: true,
			Fields: map[string]structure.StructField{
				"message": {
					Comment: "used to test a message field",
					Type:    "string",
				},
				"field": {
					Comment:  "used to test a field",
					Type:     "string",
					Optional: true,
				},
			},
		},
		"ErrorWithAllFields": {
			Comment:   "used to test a error with all fields",
			Exception: true,
			Fields: map[string]structure.StructField{
				"field": {
					Comment: "used to test a field",
					Type:    "string",
				},
				"field2": {
					Comment: "used to test a field",
					Type:    "string",
				},
			},
		},
	},
	Methods: map[string]structure.Method{
		"VoidInput": {
			Comment: "used to test a void input",
			Output:  "string",
		},
		"VoidOutput": {
			Comment:   "used to test a void output",
			Input:     "string",
			InputName: "VoidOutputInput",
		},
		"AllVoid": {
			Comment: "used to test all void",
		},
		"Cursor": {
			Comment:         "used to test a cursor",
			Output:          "string",
			OutputBehaviour: structure.OutputBehaviourCursor,
		},
		"OptionalCursor": {
			Comment:         "used to test a optional cursor",
			Output:          "string",
			OutputBehaviour: structure.OutputBehaviourCursor,
			OutputOptional:  true,
		},
		"NoComment": {
			Input:     "string",
			InputName: "NoCommentInput",
			Output:    "string",
		},
		"StructOutput": {
			Comment: "used to test a struct output",
			Output:  "OneField",
		},
		"StructOptionalOutput": {
			Comment:        "used to test a optional struct output",
			Output:         "OneField",
			OutputOptional: true,
		},
		"StructCursorOutput": {
			Comment:         "used to test a struct cursor output",
			Output:          "OneField",
			OutputBehaviour: structure.OutputBehaviourCursor,
			OutputOptional:  true,
		},
	},
}

func doCompilation(t *testing.T, name string, opts map[string]string) {
	t.Helper()
	m, err := rpc.Compile(name, filledStructure, opts)
	assert.NoError(t, err)
	for k, v := range m {
		if golden.Update() {
			golden.SetP(t, string(k), []byte(v))
		}

		valB := golden.GetP(t, string(k))
		assert.Equal(t, string(valB), v)
	}
}

func TestCompile_golang(t *testing.T) {
	tests := []struct {
		name string

		opts map[string]string
	}{
		{
			name: "default",
			opts: map[string]string{},
		},
		{
			name: "package",
			opts: map[string]string{
				"package": "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doCompilation(t, "golang", tt.opts)
		})
	}
}

func TestCompile_js(t *testing.T) {
	tests := []struct {
		name string

		opts map[string]string
	}{
		{
			name: "no esm/node",
			opts: map[string]string{
				"esm":  "false",
				"node": "false",
			},
		},
		{
			name: "esm",
			opts: map[string]string{
				"esm":  "true",
				"node": "false",
			},
		},
		{
			name: "node",
			opts: map[string]string{
				"esm":  "false",
				"node": "true",
			},
		},
		{
			name: "esm+node",
			opts: map[string]string{
				"esm":  "true",
				"node": "true",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doCompilation(t, "js", tt.opts)
		})
	}
}
