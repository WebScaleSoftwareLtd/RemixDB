// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	"crypto/sha256"
	"encoding/hex"
	goAst "go/ast"
	"reflect"
	"strings"

	"remixdb.io/ast"
	"remixdb.io/internal/engine"
)

// Builds the Go file to a string ready for prep before compilation.
func contract2go(contract *ast.ContractToken, s engine.Session) (string, error) {
	// Create the AST builder.
	builder := newAstBuilder("main")

	// Add the interface that we will use as the input.
	iface := &goAst.InterfaceType{
		Methods: &goAst.FieldList{
			List: []*goAst.Field{},
		},
	}

	// Invoke the function to build the body.
	imports, body, err := buildFunctionBody(contract, s, iface)
	if err != nil {
		return "", err
	}

	// Add the imports.
	for _, imp := range imports {
		builder.addImport(imp)
	}

	// Build the method we will call.
	builder.addFunc("Execute_hash_here", &goAst.BlockStmt{
		List: body,
	}, &goAst.FieldList{
		List: []*goAst.Field{
			{
				Names: []*goAst.Ident{
					goAst.NewIdent("r"),
				},
				Type: iface,
			},
		},
	}, &goAst.FieldList{
		List: []*goAst.Field{
			{
				Type: goAst.NewIdent("error"),
			},
		},
	})

	// Return the string.
	return builder.string()
}

// Do the compilation.
func (c *Compiler) doCompilation(contract *ast.ContractToken, s engine.Session) (reflect.Value, error) {
	// Turn the contract into a Go file.
	goFile, err := contract2go(contract, s)
	if err != nil {
		return reflect.Value{}, err
	}

	// Hash the Go file.
	shaA := sha256.Sum256([]byte(goFile))
	fileHash := hex.EncodeToString(shaA[:])

	// Turn the first usage of "Execute_hash_here" to "Execute_<hash>".
	fnName := "Execute_" + fileHash
	goFile = strings.Replace(goFile, "Execute_hash_here", fnName, 1)

	// Call the Go plugin compiler.
	plugin, err := c.GoPluginCompiler.Compile(goFile)
	if err != nil {
		return reflect.Value{}, err
	}

	// Get the function.
	fn, err := plugin.Lookup(fnName)
	if err != nil {
		return reflect.Value{}, err
	}

	// Return the reflection of the function.
	return reflect.ValueOf(fn), nil
}
