// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	"go/ast"
	"testing"

	"github.com/jimeh/go-golden"
	"github.com/stretchr/testify/assert"
)

func Test_newAstBuilder(t *testing.T) {
	a := newAstBuilder("horse")
	s, err := a.string()
	assert.NoError(t, err)
	assert.Equal(t, "package horse\n", s)
}

func Test_astBuilder_addFunc(t *testing.T) {
	tests := []struct {
		name string

		inputs  *ast.FieldList
		body    *ast.BlockStmt
		results *ast.FieldList
	}{
		{
			name: "no inputs, no results",
			body: &ast.BlockStmt{
				List: []ast.Stmt{},
			},
		},
		{
			name: "string input, no results",
			inputs: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent("input"),
						},
						Type: ast.NewIdent("string"),
					},
				},
			},
			body: &ast.BlockStmt{},
		},
		{
			name: "string input, string result",
			inputs: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent("input"),
						},
						Type: ast.NewIdent("string"),
					},
				},
			},
			body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("input"),
						},
					},
				},
			},
			results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("string"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newAstBuilder("main")
			a.addFunc("test", tt.body, tt.inputs, tt.results)
			s, err := a.string()
			if assert.NoError(t, err) && golden.Update() {
				golden.Set(t, []byte(s))
			}
			assert.Equal(t, string(golden.Get(t)), s)
		})
	}
}

func Test_astBuilder_addImport(t *testing.T) {
	tests := []struct {
		name string

		imports      []string
		beforeAction func(a *astBuilder)
	}{
		{
			name:    "one import",
			imports: []string{"fmt"},
		},
		{
			name:    "one import then duplicate",
			imports: []string{"fmt", "fmt"},
		},
		{
			name:    "two imports",
			imports: []string{"fmt", "strings"},
		},
		{
			name:    "two imports then duplicate",
			imports: []string{"fmt", "strings", "fmt"},
		},
		{
			name:    "import with function made first",
			imports: []string{"fmt"},
			beforeAction: func(a *astBuilder) {
				a.addFunc("test", &ast.BlockStmt{}, nil, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newAstBuilder("main")
			if tt.beforeAction != nil {
				tt.beforeAction(a)
			}
			for _, imp := range tt.imports {
				a.addImport(imp)
			}
			s, err := a.string()
			if assert.NoError(t, err) && golden.Update() {
				golden.Set(t, []byte(s))
			}
			assert.Equal(t, string(golden.Get(t)), s)
		})
	}
}
