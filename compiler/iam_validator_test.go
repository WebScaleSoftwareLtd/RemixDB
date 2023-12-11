// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/jimeh/go-golden"
	"github.com/stretchr/testify/assert"
)

func Test_iamValidator(t *testing.T) {
	tests := []struct {
		name string

		permissions []string
		cursor      bool
	}{
		{
			name:        "single permission no cursor",
			permissions: []string{"hello:world"},
		},
		{
			name:        "single permission cursor",
			permissions: []string{"hello:world"},
			cursor:      true,
		},
		{
			name:        "single then duplicate permission no cursor",
			permissions: []string{"hello:world", "hello:world"},
		},
		{
			name:        "single then duplicate permission cursor",
			permissions: []string{"hello:world", "hello:world"},
			cursor:      true,
		},

		{
			name:        "multiple permission no cursor",
			permissions: []string{"hello:world", "cat:dog"},
		},
		{
			name:        "multiple permission cursor",
			permissions: []string{"hello:world", "cat:dog"},
			cursor:      true,
		},
		{
			name:        "multiple then duplicate permission no cursor",
			permissions: []string{"hello:world", "cat:dog", "hello:world", "cat:dog"},
		},
		{
			name:        "multiple then duplicate permission cursor",
			permissions: []string{"hello:world", "cat:dog", "hello:world", "cat:dog"},
			cursor:      true,
		},

		{
			name:        "same namespace no cursor",
			permissions: []string{"a:b", "hello:world", "x:y", "hello:cat", "hello:dog", "hello:mouse"},
		},
		{
			name:        "same namespace cursor",
			permissions: []string{"a:b", "hello:world", "x:y", "hello:cat", "hello:dog", "hello:mouse"},
			cursor:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Defines the function statements.
			stmts := []ast.Stmt{
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: ast.NewIdent("println"),
						Args: []ast.Expr{
							&ast.BasicLit{
								Kind:  token.STRING,
								Value: `"hello world"`,
							},
						},
					},
				},
			}

			// Defines the interface.
			iFace := &ast.InterfaceType{
				Methods: &ast.FieldList{},
			}

			// Call the IAM validator.
			used := map[string]struct{}{}
			validator := &iamValidator{}
			validator.injectEntrypoint(&stmts, func(name string, fn *ast.FuncType) {
				addToInterface(used, iFace, name, fn)
			}, tt.cursor)

			// Add a panic after this.
			stmts = append(stmts, &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: ast.NewIdent("panic"),
					Args: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: `"AAAAAAAAAAAA"`,
						},
					},
				},
			})

			// Handle the permissions.
			for _, perm := range tt.permissions {
				validator.addValidator(perm)
			}

			// Do the compilation.
			validator.compile()

			// Call the builder.
			builder := newAstBuilder("test")
			builder.addFunc("testing", &ast.BlockStmt{
				List: stmts,
			}, &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent("r"),
						},
						Type: iFace,
					},
				},
			}, &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("error"),
					},
				},
			})
			s, err := builder.string()

			// Handle any errors.
			if assert.NoError(t, err) && golden.Update() {
				golden.Set(t, []byte(s))
			}
			assert.Equal(t, string(golden.Get(t)), s)
		})
	}
}
