// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	goAst "go/ast"

	"remixdb.io/ast"
	"remixdb.io/engine"
)

// Handles building the function body.
func buildFunctionBody(
	contract *ast.ContractToken, s engine.Session, iface *goAst.InterfaceType,
) (imports []string, body []goAst.Stmt, err error) {
	return []string{}, []goAst.Stmt{}, nil
}
