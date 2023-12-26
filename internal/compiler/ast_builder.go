// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

type astBuilder struct {
	f ast.File

	importDeclWritten bool
}

func (b *astBuilder) addImport(importName string) *astBuilder {
	b.f.Imports = append(b.f.Imports, &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: importName,
		},
	})

	// Add quotes to import name.
	importName = `"` + importName + `"`

	if b.importDeclWritten {
		// Get the import declaration.
		decl := b.f.Decls[0].(*ast.GenDecl)

		// Check if the import declaration already contains the import.
		for _, spec := range decl.Specs {
			if spec.(*ast.ImportSpec).Path.Value == importName {
				return b
			}
		}

		// Add the import to the import declaration.
		decl.Specs = append(decl.Specs, &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: importName,
			},
		})
	} else {
		// Write the top import declaration.
		b.f.Decls = append([]ast.Decl{
			&ast.GenDecl{
				Tok: token.IMPORT,
				Specs: []ast.Spec{
					&ast.ImportSpec{
						Path: &ast.BasicLit{
							Kind:  token.STRING,
							Value: importName,
						},
					},
				},
			},
		}, b.f.Decls...)
		b.importDeclWritten = true
	}

	return b
}

func (b *astBuilder) addFunc(
	name string, body *ast.BlockStmt, inputs *ast.FieldList,
	results *ast.FieldList,
) *astBuilder {
	if inputs == nil {
		inputs = &ast.FieldList{}
	}
	b.f.Decls = append(b.f.Decls, &ast.FuncDecl{
		Name: ast.NewIdent(name),
		Type: &ast.FuncType{
			Params:  inputs,
			Results: results,
		},
		Body: body,
	})
	return b
}

func (b *astBuilder) string() (string, error) {
	var buf strings.Builder
	if err := printer.Fprint(&buf, token.NewFileSet(), &b.f); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func newAstBuilder(packageName string) *astBuilder {
	return &astBuilder{
		f: ast.File{
			Package: 0,
			Name:    ast.NewIdent(packageName),
		},
	}
}
